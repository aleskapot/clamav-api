import {IExecuteFunctions, NodeOperationError,} from 'n8n-workflow';

interface BinaryData {
	fileName?: string;
	data: string;
}

interface ErrorResponse {
	error?: string;
	code?: number;
	message?: string;
}

interface FileProcessResult {
	fileName: string;
	fileBuffer: Buffer;
	formData: FormData;
}

export async function processFileForUpload(
	this: IExecuteFunctions,
	itemIndex: number,
	binaryPropertyName: string,
	customFileName: string,
	binaryData: BinaryData | undefined,
): Promise<FileProcessResult> {
	if (!binaryData) {
		// noinspection ExceptionCaughtLocallyJS
		throw new NodeOperationError(
			this.getNode(),
			`No binary data found in field "${binaryPropertyName}"`,
			{ itemIndex },
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

	return {
		fileName,
		fileBuffer,
		formData,
	};
}

export async function makeApiRequest(
	url: string,
	apiKey: string,
	formData: FormData,
): Promise<any> {
	return await fetch(url, {
		method: 'POST',
		headers: {
			'API-Key': apiKey,
		},
		body: formData,
	});
}

export function handleApiError(
	response: any,
	errorBody: ErrorResponse,
	itemIndex: number,
	node: any,
): NodeOperationError {
	// noinspection ExceptionCaughtLocallyJS
	return new NodeOperationError(
		node,
		errorBody.error
			? `[${response.status}] ${errorBody.error}: ${errorBody.message}`
			: errorBody.message || `API request failed with status ${response.status}`,
		{ itemIndex },
	);
}

export function handleProcessingError(
	error: any,
	results: any[],
	continueOnFail: () => boolean,
): boolean {
	if (continueOnFail()) {
		results.push({
			json: {
				error: error instanceof Error ? error.message : String(error),
			},
		});
		return true; // Continue processing
	}
	return false; // Re-throw error
}
