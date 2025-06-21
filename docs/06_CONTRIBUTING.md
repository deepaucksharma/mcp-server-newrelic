# Contributing to New Relic MCP Server

Welcome! We're excited that you're interested in contributing to the New Relic MCP Server project. This guide will help you get started.

## Understanding the Project

Before contributing, please read:
1. [01_VISION_AND_PHILOSOPHY.md](01_VISION_AND_PHILOSOPHY.md) - Understand why we're building this
2. [02_ARCHITECTURE.md](02_ARCHITECTURE.md) - Learn the technical approach
3. [05_ROADMAP.md](05_ROADMAP.md) - See what needs to be built

## Getting Started

### Prerequisites

- Node.js 18+ LTS
- npm, yarn, or pnpm package manager
- TypeScript knowledge
- VS Code or similar TypeScript IDE
- Git
- A New Relic account (free tier is fine)
- Claude Desktop (for testing MCP integration)

### Development Setup

1. **Fork and Clone**
   ```bash
   git clone https://github.com/YOUR-USERNAME/mcp-server-newrelic
   cd mcp-server-newrelic
   ```

2. **Create Development Branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

3. **Set Up Environment**
   ```bash
   cp .env.example .env
   # Edit .env with your New Relic credentials
   ```

4. **Install Dependencies**
   ```bash
   npm install
   # or
   yarn install
   # or
   pnpm install
   ```

5. **Build the Project**
   ```bash
   npm run build
   ```

6. **Run Tests**
   ```bash
   npm test
   ```

## Development Workflow

### 1. Pick a Task

Check the [Roadmap](05_ROADMAP.md) for current implementation phases:
- **Phase 1**: Foundation - Core discovery engine and basic tools
- **Phase 2**: Intelligence - Adaptive Query Builder and analysis
- **Phase 3**: Workflows - Workflow engine and automation
- **Phase 4**: Production - Performance and deployment

Or browse [GitHub Issues](https://github.com/deepaucksharma/mcp-server-newrelic/issues) for specific tasks.

### 2. Understand the Specification

For tool implementations, refer to [03_TOOL_SPECIFICATION.md](03_TOOL_SPECIFICATION.md) which provides:
- Input/output schemas
- Implementation hints
- Acceptance criteria

### 3. Write Code

Follow these principles:

#### Discovery-First
```typescript
// ❌ Don't assume
const badExample = (): string => {
    return "SELECT appName FROM Transaction";
};

// ✅ Always discover
const goodExample = async (): Promise<string> => {
    const fields = await discovery.getFields("Transaction");
    return buildQueryWithFields(fields);
};
```

#### Test-Driven Development
1. Write tests first
2. Implement minimal code to pass
3. Refactor for clarity

#### TypeScript Best Practices
- Interface-first design - define clear contracts
- Proper typing - avoid `any`, use strict TypeScript
- Async/await patterns - handle promises correctly
- ESLint/Prettier compliance - maintain consistent code style
- Use meaningful variable names
- Add JSDoc comments for complex logic
- Keep functions small and focused

#### Discovery-First Development Patterns

1. **Always Start with Discovery**:
   - Every tool should check if discovery data exists
   - Build world model before making assumptions
   - Use discovered fields in queries

2. **World Model Integration**:
   - Access via `ctx.worldModel` in tool handlers
   - Check confidence scores: `if (worldModel.confidence > 0.8)`
   - Update world model with new discoveries

3. **Explainability Throughout**:
   - Add trace steps: `ctx.explainabilityTrace.addStep(...)`
   - Include confidence scores in responses
   - Document discovery decisions

4. **Adaptive Query Pattern**:
   - Use Adaptive Query Builder for all NRQL generation
   - Let AQB handle field discovery and optimization
   - Never hard-code field names

5. **Error Handling with Discovery**:
   - When queries fail, attempt re-discovery
   - Provide helpful error messages about missing fields
   - Suggest alternatives based on discovery

### 4. Testing Requirements

#### Unit Tests
- Minimum 80% coverage
- Test both success and error cases
- Use Jest/Vitest patterns with describe/it blocks

```typescript
describe('DiscoveryTool', () => {
    const testCases = [
        {
            name: 'should discover event types successfully',
            input: { accountId: '123' } as DiscoveryRequest,
            expected: { eventTypes: ['Transaction'] } as DiscoveryResponse,
            shouldThrow: false,
        },
        // More test cases...
    ];
    
    testCases.forEach(({ name, input, expected, shouldThrow }) => {
        it(name, async () => {
            if (shouldThrow) {
                await expect(discoveryTool.execute(input)).rejects.toThrow();
            } else {
                const result = await discoveryTool.execute(input);
                expect(result).toEqual(expected);
            }
        });
    });
});
```

#### Integration Tests
- Test against mock New Relic responses
- Validate complete workflows
- Ensure proper error handling
- Use MSW (Mock Service Worker) for API mocking
- Test async operations with proper awaiting

### 5. Documentation

Update documentation when you:
- Add new features
- Change existing behavior
- Discover important implementation details
- Find useful patterns

## Submission Process

### 1. Pre-Submission Checklist

- [ ] Code follows project style guidelines
- [ ] All tests pass (`npm test`)
- [ ] TypeScript compilation succeeds (`npm run build`)
- [ ] ESLint checks pass (`npm run lint`)
- [ ] New tests added for new functionality
- [ ] Documentation updated if needed
- [ ] Commit messages are descriptive
- [ ] Branch is up to date with main

### 2. Commit Messages

Follow conventional commits:
```
feat: add discovery.explore_event_types tool implementation
fix: handle empty response in TypeScript query tool
docs: update roadmap with completed TypeScript tasks
test: add Jest integration tests for discovery workflow
refactor: migrate Go implementation to TypeScript
```

### 3. Pull Request

1. **Create PR** with descriptive title
2. **Fill out PR template** (if provided)
3. **Link related issues** using "Fixes #123"
4. **Describe changes** and testing approach
5. **Update roadmap** if completing a task

### 4. Code Review

Be prepared to:
- Explain your design decisions
- Make requested changes
- Add additional tests if needed
- Discuss alternative approaches

## Architecture Guidelines

### Tool Implementation Pattern

```typescript
import { Tool, ToolMetadata, PriorityLevel } from '@modelcontextprotocol/sdk/types.js';
import { DiscoveryClient } from '../discovery/client.js';

interface DiscoveryRequest {
    accountId: string;
    limit?: number;
}

interface DiscoveryResponse {
    eventTypes: string[];
    metadata: {
        totalFound: number;
        executionTime: number;
    };
}

// DiscoveryExploreEventTypes implements the discovery.explore_event_types tool
export class DiscoveryExploreEventTypes implements Tool {
    constructor(private client: DiscoveryClient) {}

    // Metadata returns tool metadata for registration
    get metadata(): ToolMetadata {
        return {
            name: 'discovery.explore_event_types',
            description: 'Discover available event types in the account',
            category: 'discovery',
            priority: PriorityLevel.High,
            inputSchema: {
                type: 'object',
                properties: {
                    accountId: { type: 'string', description: 'New Relic account ID' },
                    limit: { type: 'number', description: 'Maximum results to return', default: 100 }
                },
                required: ['accountId']
            }
        };
    }

    // Execute runs the tool with given parameters
    async execute(params: DiscoveryRequest): Promise<DiscoveryResponse> {
        try {
            // 1. Validate parameters
            this.validateParams(params);
            
            // 2. Call discovery engine
            const eventTypes = await this.client.discoverEventTypes(params.accountId, params.limit);
            
            // 3. Format response
            return {
                eventTypes,
                metadata: {
                    totalFound: eventTypes.length,
                    executionTime: Date.now()
                }
            };
        } catch (error) {
            // 4. Handle errors gracefully
            throw new Error(`Failed to discover event types: ${error instanceof Error ? error.message : 'Unknown error'}`);
        }
    }

    private validateParams(params: DiscoveryRequest): void {
        if (!params.accountId || typeof params.accountId !== 'string') {
            throw new Error('Valid accountId is required');
        }
    }
}
```

### Error Handling

Always provide helpful error messages:

```typescript
// ❌ Bad
throw new Error('failed');

// ✅ Good
throw new Error(`Failed to discover event types: ${error instanceof Error ? error.message : 'Unknown error'}`);

// ✅ Even better - custom error types
class DiscoveryError extends Error {
    constructor(
        message: string,
        public readonly code: string,
        public readonly cause?: Error
    ) {
        super(message);
        this.name = 'DiscoveryError';
    }
}

throw new DiscoveryError(
    'Failed to discover event types',
    'DISCOVERY_FAILED',
    originalError
);
```

### Performance Considerations

- Use AbortController for cancellation
- Implement reasonable timeouts with Promise.race
- Cache discovery results appropriately using Map or LRU cache
- Batch operations where possible using Promise.all
- Use proper TypeScript async/await patterns
- Consider memory usage with large result sets

## Development Tips

### Local Testing with Claude Desktop

1. Build the server:
   ```bash
   npm run build
   ```

2. Configure Claude Desktop to use your local build

3. Test your changes interactively

### Debugging

Use structured logging:
```typescript
import { logger } from '../utils/logger.js';

logger.debug('Executing tool', {
    tool: 'discovery.explore_event_types',
    params,
    timestamp: new Date().toISOString()
});

// Or with a more structured approach
const log = logger.child({ tool: 'discovery.explore_event_types' });
log.debug('Executing tool', { params });
```

### Common Pitfalls

1. **Assuming Field Names**: Always discover, never assume
2. **Ignoring Errors**: Handle all error cases explicitly
3. **Missing Tests**: Write tests for edge cases
4. **Poor Documentation**: Document "why" not just "what"
5. **Using `any` Type**: Always provide proper TypeScript types
6. **Forgetting `await`**: Always await async operations
7. **Not Handling Promise Rejections**: Use try/catch blocks
8. **Improper Interface Design**: Define clear contracts upfront

## Getting Help

- 💬 **Discussions**: Use GitHub Discussions for questions
- 🐛 **Issues**: Report bugs with reproducible examples
- 💡 **Ideas**: Propose new features with use cases
- 📚 **Docs**: Contribute to documentation improvements

## Recognition

Contributors will be:
- Listed in the project README
- Credited in release notes
- Invited to project planning discussions

## Code of Conduct

We are committed to providing a welcoming and inclusive environment. Please:
- Be respectful and constructive
- Focus on what's best for the project
- Show empathy towards others
- Accept feedback gracefully

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing to the future of discovery-first observability! 🚀