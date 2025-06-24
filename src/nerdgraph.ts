import fetch from 'node-fetch';

export class NerdGraphClient {
  private apiKey: string;
  private endpoint = 'https://api.newrelic.com/graphql';

  constructor(apiKey: string) {
    this.apiKey = apiKey;
  }

  async query(query: string, variables?: Record<string, any>): Promise<any> {
    const response = await fetch(this.endpoint, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'API-Key': this.apiKey,
      },
      body: JSON.stringify({
        query,
        variables,
      }),
    });

    if (!response.ok) {
      throw new Error(`NerdGraph API error: ${response.statusText}`);
    }

    const data = await response.json() as any;
    
    if (data.errors) {
      throw new Error(`GraphQL errors: ${JSON.stringify(data.errors)}`);
    }

    return data.data;
  }
}