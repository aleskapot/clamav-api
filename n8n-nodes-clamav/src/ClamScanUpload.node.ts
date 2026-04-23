import {
	IExecuteFunctions,
	INodeExecutionData,
	INodeType,
	INodeTypeDescription,
	NodeConnectionTypes,
	NodeOperationError,
} from 'n8n-workflow';

interface UploadResponse {
	file_id: string;
	file_name: string;
	file_size: number;
	message: string;
	received_at: string;
}

interface ErrorResponse {
	error?: string;
	code?: number;
	message?: string;
}

export class ClamScanUpload implements INodeType {
	description: INodeTypeDescription = {
		displayName: 'ClamAV File Upload',
		name: 'clamScanUpload',
		icon: 'file:clamav.svg',
		group: ['transform'],
		version: 1,
		description: 'Asynchronously upload files for virus scanning with webhook notification',
		subtitle: '={{$parameter["fileName"] || "from input"}}',
		defaults: {
			name: 'ClamAV File Upload',
		},
		inputs: [{ type: NodeConnectionTypes.Main, displayName: 'Input' }],
		outputs: [{ type: NodeConnectionTypes.Main, displayName: 'Output' }],
		outputNames: ['Output'],
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
				description: 'The name of the incoming binary field containing the file to upload',
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

		const results: INodeExecutionData[] = [];

		for (let i = 0; i < items.length; i++) {
			try {
				const binaryPropertyName = this.getNodeParameter('binaryPropertyName', i) as string;
				const customFileName = this.getNodeParameter('fileName', i) as string;

				const binaryData = items[i].binary?.[binaryPropertyName];

				if (!binaryData) {
					// noinspection ExceptionCaughtLocallyJS
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

				const response = await fetch(`${credentials.apiUrl}/files/upload`, {
					method: 'POST',
					headers: {
						'API-Key': credentials.apiKey,
					},
					body: formData,
				});

				if (!response.ok) {
					const errorBody = (await response.json().catch(() => ({}))) as ErrorResponse;
					// noinspection ExceptionCaughtLocallyJS
					throw new NodeOperationError(
						this.getNode(),
						errorBody.error
							? `[${response.status}] ${errorBody.error}: ${errorBody.message}`
							: errorBody.message || `API request failed with status ${response.status}`,
						{ itemIndex: i },
					);
				}

				const uploadResult = (await response.json()) as UploadResponse;

				const requestId = response.headers.get('Request-Id') || uploadResult.file_id;

				results.push({
					json: {
						requestId,
						fileId: uploadResult.file_id,
						fileName: uploadResult.file_name,
						fileSize: uploadResult.file_size,
						message: uploadResult.message,
						receivedAt: uploadResult.received_at,
					},
				});
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