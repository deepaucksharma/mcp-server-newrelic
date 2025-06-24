import { NerdGraphClient } from '../nerdgraph.js';

export class HelloTool {
  constructor(private nerdgraph: NerdGraphClient) {}

  getDefinition() {
    return {
      name: 'hello_newrelic',
      description: 'Test connection to New Relic and return account information',
      inputSchema: {
        type: 'object',
        properties: {
          account_id: {
            type: 'number',
            description: 'New Relic account ID to query',
          },
        },
        required: ['account_id'],
      },
    };
  }

  async execute(args: any) {
    const { account_id } = args;

    // Simple GraphQL query to get account info
    const query = `
      query GetAccountInfo($accountId: Int!) {
        actor {
          account(id: $accountId) {
            id
            name
            nrql(query: "SELECT count(*) FROM Transaction SINCE 1 hour ago") {
              results
            }
          }
        }
      }
    `;

    try {
      const result = await this.nerdgraph.query(query, { accountId: account_id });
      
      const account = result.actor.account;
      const transactionCount = account.nrql.results[0]?.count || 0;

      return {
        content: [
          {
            type: 'text',
            text: `Hello from New Relic! 👋\n\nAccount Information:\n- ID: ${account.id}\n- Name: ${account.name}\n- Transactions (last hour): ${transactionCount}`,
          },
        ],
      };
    } catch (error) {
      return {
        content: [
          {
            type: 'text',
            text: `Error connecting to New Relic: ${error instanceof Error ? error.message : String(error)}`,
          },
        ],
        isError: true,
      };
    }
  }
}