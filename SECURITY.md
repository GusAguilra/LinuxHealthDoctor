# Security Policy

## Supported Versions

| Version | Supported |
|---|---|
| >= 1.0.0 | Yes |
| < 1.0.0 (alpha/beta) | No |

## Reporting a Vulnerability

Linux Health Doctor takes security seriously. If you discover a security vulnerability, please report it privately.

**Do not report security vulnerabilities through public GitHub issues.**

Please email security@linuxhealthdoctor.dev (placeholder) with details including:
- Type of issue
- Full description
- Steps to reproduce
- Affected versions
- Any potential impacts

You should receive a response within 48 hours. We will keep you informed as we work on a fix.

## Security Design

- **No telemetry**: Zero data leaves your machine under any circumstances
- **No cloud dependency**: Fully offline, no external service calls
- **Deterministic analysis**: Rule-based engine, no external AI/ML models
- **Single binary**: No runtime dependencies, minimal attack surface
- **Least privilege**: Runs at user level by default; sudo only when required
- **Custom check sandboxing**: User-defined checks run in restricted environments
- **Signed releases**: All releases signed with Cosign
- **SBOM**: Software Bill of Materials generated per release

## Supply Chain Security

- Dependencies pinned to specific versions
- `go.sum` for integrity verification
- Reproducible builds with Go's buildid
- Dependabot monitoring for dependency vulnerabilities
- Minimal dependency surface (pure Go libraries preferred)
