import {
	ICredentialType,
	INodeProperties,
	Icon,
} from 'n8n-workflow';

export class ClamScanApi implements ICredentialType {
	name = 'clamScanApi';

	displayName = 'ClamAV Virus Scanner';

	icon: Icon = 'file:clamav.svg';

	documentationUrl = 'https://github.com/aleskapot/clamav-api';

	properties: INodeProperties[] = [
		{
			displayName: 'API URL',
			name: 'apiUrl',
			type: 'string',
			default: 'http://localhost:8080',
			placeholder: 'http://localhost:8080',
			description: 'Base URL of the ClamAV API server',
		},
		{
			displayName: 'API Key',
			name: 'apiKey',
			type: 'string',
			typeOptions: {
				password: true,
			},
			default: '',
			description: 'API key for authentication',
		},
	];
}
