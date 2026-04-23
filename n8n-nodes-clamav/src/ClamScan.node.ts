import {
	IExecuteFunctions,
	INodeExecutionData,
	INodeType,
	INodeTypeDescription,
	NodeConnectionTypes,
	NodeOperationError,
} from 'n8n-workflow';

interface ScanResponse {
	file_id: string;
	file_name: string;
	file_size: number;
	result: string;
	threat?: string;
	duration_ms: number;
	scanned_at: string;
}

interface ErrorResponse {
	message?: string;
}

export class ClamScan implements INodeType {
	description: INodeTypeDescription = {
		displayName: 'ClamAV File Scan',
		name: 'clamScan',
		icon: 'file:clamav.svg',
		group: ['transform'],
		version: 1,
		description: 'Scan files for viruses using ClamAV',
		subtitle: '={{$parameter["fileName"] || "from input"}}',
		defaults: {
			name: 'ClamAV File Scan',
		},
		inputs: [{ type: NodeConnectionTypes.Main, displayName: 'Input' }],
		outputs: [{ type: NodeConnectionTypes.Main, displayName: 'Output' }],
		credentials: [
			{
				name: 'clamScanApi',
				required: true,
			},
		],
		properties: [
			{
				displayName: 'Operation',
				name: 'operation',
				type: 'options',
				options: [
					{
						name: 'Scan File',
						value: 'scanFile',
						description: 'Synchronously scan a file for viruses',
					},
				],
				default: 'scanFile',
			},
			{
				displayName: 'Input Binary Field',
				name: 'binaryPropertyName',
				type: 'string',
				default: 'data',
				required: true,
				displayOptions: {
					show: {
						operation: ['scanFile'],
					},
				},
				description: 'The name of the incoming binary field containing the file to scan',
			},
			{
				displayName: 'File Name',
				name: 'fileName',
				type: 'string',
				default: '',
				required: false,
				displayOptions: {
					show: {
						operation: ['scanFile'],
					},
				},
				description: 'Optional file name. If not provided, will use the binary file name',
			},
		],
	};

	async execute(this: IExecuteFunctions): Promise<INodeExecutionData[][]> {
		const items = this.getInputData();
		const operation = this.getNodeParameter('operation', 0) as string;

		const credentials = await this.getCredentials<{
			apiUrl: string;
			apiKey: string;
		}>('clamScanApi');

		const results: INodeExecutionData[] = [];

		for (let i = 0; i < items.length; i++) {
			try {
				if (operation === 'scanFile') {
					const binaryPropertyName = this.getNodeParameter('binaryPropertyName', i) as string;
					const customFileName = this.getNodeParameter('fileName', i) as string;

					const binaryData = items[i].binary?.[binaryPropertyName];

					if (!binaryData) {
						throw new NodeOperationError(
							this.getNode(),
							`No binary data found in field "${binaryPropertyName}"`,
							{ itemIndex: i },
						);
					}

					const fileName = customFileName || binaryData.fileName || 'unknown.bin';
					const fileBuffer = Buffer.from(binaryData.data, 'base64');

					const formData = new FormData();
					formData.append(
						'file',
						new Blob([fileBuffer]),
						fileName,
					);

					const response = await fetch(`${credentials.apiUrl}/files/scan`, {
						method: 'POST',
						headers: {
							'API-Key': credentials.apiKey,
						},
						body: formData,
					});

					if (!response.ok) {
						const errorBody = (await response.json().catch(() => ({}))) as ErrorResponse;
						throw new NodeOperationError(
							this.getNode(),
							errorBody.message || `API request failed with status ${response.status}`,
							{ itemIndex: i },
						);
					}

					const scanResult = (await response.json()) as ScanResponse;

					results.push({
						json: {
							fileId: scanResult.file_id,
							fileName: scanResult.file_name,
							fileSize: scanResult.file_size,
							result: scanResult.result,
							threat: scanResult.threat || null,
							durationMs: scanResult.duration_ms,
							scannedAt: scanResult.scanned_at,
						},
					});
				}
			} catch (error) {
				if (this.continueOnFail()) {
					results.push({
						json: {
							error: error instanceof Error ? error.message : String(error),
						},
					});
					continue;
				}
				throw error;
			}
		}

		return [results];
	}
}
