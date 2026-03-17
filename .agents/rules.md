# Project Rules for FireboxAuditor

This project is a security auditing tool for WatchGuard Firebox configurations.

## Architecture

- **Backend**: Written in Go.
  - `audit.go`: Core auditing logic.
  - `parser.go`: XML parsing logic for Firebox config files.
  - `ssh_client.go`: Handles SSH connections to Firebox devices.
  - `main.go`: Application entry point.
- **Frontend**: Located in the `frontend/` directory.
  - Built with React, Vite, and Tailwind CSS.
  - Follows WatchGuard Brand Guidelines (Typography: Inter/Roboto, Colors: WatchGuard Red, Black, Gray).

## Coding Standards

### Go (Backend)
- Use standard library for XML processing where possible.
- Error handling: Return errors upwards; use `fmt.Errorf` with wrapping for context.
- Keep `audit.go` and `parser.go` clean and well-documented.

### React (Frontend)
- Use functional components and hooks.
- Styling: Use Tailwind CSS classes for layout and brand colors.
- i18n: All user-facing strings must be translated using the project's i18n system.

### SSH & Security
- SSH connections must be handled securely.
- Avoid logging sensitive information like feature keys or passwords.

## Task Specifics
- When adding new audit rules, update both the backend logic and ensure the frontend can visualize the new rule results.
- Always verify UI changes against the WatchGuard Brand Guidelines.
