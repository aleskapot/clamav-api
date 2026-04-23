import {
	IExecuteFunctions,
	INodeExecutionData,
	INodeType,
	INodeTypeDescription,
	NodeConnectionTypes,
} from 'n8n-workflow';
import {
	processFileForUpload,
	makeApiRequest,
	handleApiError,
	handleProcessingError,
} from './utils';

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
	error?: string;
	code?: number;
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
		outputs: [
			{ type: NodeConnectionTypes.Main, displayName: 'Clean' },
			{ type: NodeConnectionTypes.Main, displayName: 'Infected' },
		],
		outputNames: ['Clean', 'Infected'],
		credentials: [
			{
				name: 'clamScanApi',
				required: true,
			},
		],
		properties: [
			{
				displayName: 'Input Binary Field',
				name: 'binaryPropertyName',
				type: 'string',
				default: 'data',
				required: true,
				description: 'The name of the incoming binary field containing the file to scan',
			},
			{
				displayName: 'File Name',
				name: 'fileName',
				type: 'string',
				default: '',
				required: false,
				description: 'Optional file name. If not provided, will use the binary file name',
			},
		],
	};

	async execute(this: IExecuteFunctions): Promise<INodeExecutionData[][]> {
		const items = this.getInputData();

		const credentials = await this.getCredentials<{
			apiUrl: string;
			apiKey: string;
		}>('clamScanApi');

		const cleanResults: INodeExecutionData[] = [];
		const infectedResults: INodeExecutionData[] = [];

		for (let i = 0; i < items.length; i++) {
			try {
				const binaryPropertyName = this.getNodeParameter('binaryPropertyName', i) as string;
				const customFileName = this.getNodeParameter('fileName', i) as string;

				const binaryData = items[i].binary?.[binaryPropertyName];

				const { formData } = await processFileForUpload.call(
					this,
					i,
					binaryPropertyName,
					customFileName,
					binaryData,
				);

				const response = await makeApiRequest(
					`${credentials.apiUrl}/files/scan`,
					credentials.apiKey,
					formData,
				);

				if (!response.ok) {
					const errorBody = (await response.json().catch(() => ({}))) as ErrorResponse;
					// noinspection ExceptionCaughtLocallyJS
					throw handleApiError(response, errorBody, i, this.getNode());
				}

				const scanResult = (await response.json()) as ScanResponse;

				const requestId = response.headers.get('Request-Id') || scanResult.file_id;

				const metadata = {
					requestId,
					fileId: scanResult.file_id,
					fileName: scanResult.file_name,
					fileSize: scanResult.file_size,
					durationMs: scanResult.duration_ms,
				};

				const isClean = scanResult.result === 'clean';

				if (isClean) {
					cleanResults.push({
						json: {
							...metadata,
						},
					});
				} else {
					infectedResults.push({
						json: {
							...metadata,
							threat: scanResult.threat || scanResult.result,
						},
					});
				}
			} catch (error) {
				if (handleProcessingError(error, cleanResults, () => this.continueOnFail())) {
					continue;
				}
				throw error;
			}
		}

		return [cleanResults, infectedResults];
	}
}