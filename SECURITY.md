# Security Policy

## Reporting Security Vulnerabilities

We take security seriously. If you discover a security vulnerability, please follow responsible disclosure practices:

1. **DO NOT** create a public GitHub issue
2. Email security concerns to: [security@deepaucksharma.com] (replace with actual contact)
3. Include:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if any)

We will acknowledge receipt within 48 hours and provide updates on the resolution.

## Security Features

### Authentication & Authorization
- **API Key Validation**: All New Relic API keys are validated before use
- **Multi-Account Isolation**: Cross-account access requires explicit configuration
- **Session Management**: Secure session handling with configurable TTLs
- **No Default Credentials**: No hardcoded or default authentication values

### Data Protection
- **No Credential Logging**: API keys and sensitive data are never logged
- **Input Validation**: All user inputs are validated and sanitized
- **Query Validation**: NRQL queries are validated before execution
- **Rate Limiting**: Built-in rate limiting prevents abuse
- **Timeout Controls**: All operations have configurable timeouts

### Tool Safety Levels
Each tool is classified by safety level:
- **Safe**: Read-only operations (all discovery and query tools)
- **Caution**: Operations that might impact performance
- **Destructive**: Operations that modify data (require confirmation)

### Operational Security
- **Dry-Run Mode**: Test mutations without making changes
- **Audit Logging**: All destructive operations are logged
- **Circuit Breakers**: Prevent cascading failures
- **Graceful Degradation**: Falls back safely on errors

## Best Practices

### For Developers
1. **Never commit credentials** - Use environment variables
2. **Validate all inputs** - Especially in tool handlers
3. **Use least privilege** - Request minimal New Relic permissions
4. **Test in mock mode** - Develop without real API access
5. **Review tool safety** - Classify new tools appropriately

### For Users
1. **Secure your API keys** - Treat them like passwords
2. **Use read-only keys** when possible
3. **Monitor API usage** - Check New Relic API key usage regularly
4. **Limit account access** - Only grant necessary permissions
5. **Review audit logs** - Check for unexpected operations

### For Deployment
1. **Use HTTPS** for HTTP transport mode
2. **Enable TLS** for Redis connections
3. **Restrict network access** - Use firewalls/security groups
4. **Monitor the server** - Use APM to track behavior
5. **Keep updated** - Apply security patches promptly

## Security Checklist

- [ ] API keys stored in environment variables
- [ ] No credentials in code or logs
- [ ] Input validation implemented
- [ ] Rate limiting configured
- [ ] Timeouts set appropriately
- [ ] Mock mode used for development
- [ ] HTTPS enabled for production
- [ ] Network access restricted
- [ ] Monitoring configured
- [ ] Audit logging enabled

## Known Limitations

1. **EU Region**: Not yet fully supported (planned)
2. **Fine-grained RBAC**: Currently uses New Relic's permission model
3. **Token Rotation**: Manual API key rotation required

## Security Updates

Security patches are released as soon as possible after discovery. To stay updated:

1. Watch the repository for releases
2. Enable security alerts in GitHub
3. Subscribe to release notifications
4. Review CHANGELOG for security fixes

## Compliance

The MCP Server is designed to help with:
- **Data Residency**: Respects New Relic region settings
- **Audit Requirements**: Comprehensive operation logging
- **Access Control**: Leverages New Relic's IAM

Note: This software is provided as-is. Users are responsible for compliance with their specific regulatory requirements.