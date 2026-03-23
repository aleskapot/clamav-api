package clamscan

import (
	"bufio"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/clamav-api/internal/config"
	"github.com/clamav-api/internal/model"
)

type Client struct {
	address string
	timeout time.Duration
}

func NewClient(cfg *config.ClamAVConfig) *Client {
	return &Client{
		address: cfg.Address(),
		timeout: cfg.Timeout,
	}
}

func (c *Client) GetTimeout() time.Duration {
	return c.timeout
}

func (c *Client) Ping(ctx context.Context) error {
	conn, err := c.dialContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to clamav: %w", err)
	}
	defer conn.Close()

	if err := c.sendCommand(conn, "nPING"); err != nil {
		return fmt.Errorf("failed to send PING: %w", err)
	}

	resp, err := c.readResponse(conn)
	if err != nil {
		return fmt.Errorf("failed to read PING response: %w", err)
	}

	if resp != "PONG" {
		return fmt.Errorf("unexpected PING response: %s", resp)
	}

	return nil
}

func (c *Client) Version(ctx context.Context) (string, error) {
	conn, err := c.dialContext(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to connect to clamav: %w", err)
	}
	defer conn.Close()

	if err := c.sendCommand(conn, "nVERSION"); err != nil {
		return "", fmt.Errorf("failed to send VERSION: %w", err)
	}

	resp, err := c.readResponse(conn)
	if err != nil {
		return "", fmt.Errorf("failed to read VERSION response: %w", err)
	}

	return resp, nil
}

func (c *Client) Stats(ctx context.Context) (string, error) {
	conn, err := c.dialContext(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to connect to clamav: %w", err)
	}
	defer conn.Close()

	if err := c.sendCommand(conn, "nSTATS"); err != nil {
		return "", fmt.Errorf("failed to send STATS: %w", err)
	}

	resp, err := c.readAllLines(conn)
	if err != nil {
		return "", fmt.Errorf("failed to read STATS response: %w", err)
	}

	return resp, nil
}

func (c *Client) ScanStream(ctx context.Context, data io.Reader) (*model.ScanResponse, time.Time, error) {
	conn, err := c.dialContext(ctx)
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("failed to connect to clamav: %w", err)
	}

	if err := c.sendCommand(conn, "nINSTREAM"); err != nil {
		conn.Close()
		return nil, time.Time{}, fmt.Errorf("failed to send INSTREAM: %w", err)
	}

	startTime := time.Now()

	buf := make([]byte, 64*1024)

	for {
		n, err := data.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			conn.Close()
			return nil, time.Time{}, fmt.Errorf("failed to read data: %w", err)
		}

		chunk := buf[:n]

		sizeBuf := make([]byte, 4)
		binary.BigEndian.PutUint32(sizeBuf, uint32(len(chunk)))

		if _, err := conn.Write(sizeBuf); err != nil {
			conn.Close()
			return nil, time.Time{}, fmt.Errorf("failed to send chunk size: %w", err)
		}

		if _, err := conn.Write(chunk); err != nil {
			conn.Close()
			return nil, time.Time{}, fmt.Errorf("failed to send chunk: %w", err)
		}

		select {
		case <-ctx.Done():
			conn.Close()
			return nil, time.Time{}, ctx.Err()
		default:
		}
	}

	terminator := make([]byte, 4)
	if _, err := conn.Write(terminator); err != nil {
		conn.Close()
		return nil, time.Time{}, fmt.Errorf("failed to send terminator: %w", err)
	}

	if err := conn.SetReadDeadline(time.Now().Add(c.timeout)); err != nil {
		return nil, time.Time{}, err
	}

	resp, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		conn.Close()
		if isTimeout(err) {
			return nil, time.Time{}, fmt.Errorf("scan timeout")
		}
		return nil, time.Time{}, fmt.Errorf("failed to read response: %w", err)
	}

	conn.Close()
	resp = strings.TrimSpace(resp)

	duration := time.Since(startTime)

	result := &model.ScanResponse{
		DurationMs: duration.Milliseconds(),
		ScannedAt:  startTime,
	}

	if strings.HasPrefix(resp, "stream: OK") {
		result.Result = model.ResultClean
	} else if strings.Contains(resp, "FOUND") {
		result.Result = model.ResultInfected
		result.Threat = c.extractThreat(resp)
	} else if strings.Contains(resp, "ERROR") || strings.Contains(resp, "FAIL") {
		result.Result = model.ResultError
		result.Threat = resp
	} else {
		result.Result = model.ResultError
		result.Threat = resp
	}

	return result, startTime, nil
}

func (c *Client) dialContext(ctx context.Context) (net.Conn, error) {
	dialer := &net.Dialer{
		Timeout: c.timeout,
	}

	conn, err := dialer.DialContext(ctx, "tcp", c.address)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (c *Client) sendCommand(conn net.Conn, cmd string) error {
	_, err := conn.Write([]byte(cmd + "\n"))
	return err
}

func (c *Client) readResponse(conn net.Conn) (string, error) {
	err := conn.SetReadDeadline(time.Now().Add(c.timeout))
	if err != nil {
		return "", err
	}
	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func (c *Client) readAllLines(conn net.Conn) (string, error) {
	err := conn.SetReadDeadline(time.Now().Add(c.timeout))
	if err != nil {
		return "", err
	}
	reader := bufio.NewReader(conn)
	var lines []string
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			if len(line) > 0 {
				lines = append(lines, strings.TrimSpace(line))
			}
			break
		}
		if err != nil {
			return "", err
		}
		lines = append(lines, strings.TrimSpace(line))
	}
	return strings.Join(lines, "\n"), nil
}

func (c *Client) extractThreat(resp string) string {
	parts := strings.Split(resp, " ")
	if len(parts) >= 2 {
		threat := parts[1]
		threat = strings.TrimSuffix(threat, "FOUND")
		return strings.TrimSpace(threat)
	}
	return resp
}

func isTimeout(err error) bool {
	if netErr, ok := errors.AsType[net.Error](err); ok {
		return netErr.Timeout()
	}
	return false
}
