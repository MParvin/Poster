# Agent Instructions for Social Poster Project

## Overview

This document provides guidance for AI agents working on the Social Poster project.

## Development Principles

1.  **Modularity**: Keep functionalities well-separated, especially for different social media platforms and core logic (config, GitHub interaction, posting).
2.  **Security**:
    *   Credentials must always be loaded from environment variables (via `.env` file or system environment). Do not hardcode secrets.
    *   When logging, be careful not to expose full tokens or sensitive information. Use truncation or masking if necessary (as seen in `posting` package placeholders).
3.  **Configuration**: All configurable parameters (API keys, URLs, paths) should be managed through `config/config.go` and populated from environment variables.
4.  **Error Handling**: Implement robust error handling. Log errors clearly and allow the application to fail gracefully or retry where appropriate (though retry logic is not yet implemented).
5.  **Dependencies**: Use Go modules for dependency management. Keep dependencies to a minimum where practical.
6.  **Git Usage**:
    *   The current GitHub integration shells out to the `git` command. Ensure this is handled safely.
    *   Future improvements might involve using a native Go git library for better control and security.

## Code Structure

-   `cmd/`: Cobra CLI commands.
-   `config/`: Configuration loading.
-   `github/`: Interacting with the posts GitHub repository.
-   `posting/`: Logic for posting to different social media platforms. Each platform should have its own file.
-   `main.go`: Main application entry point.
-   `Dockerfile` & `docker-compose.yml`: For containerization and orchestration.

## Tasks & Workflow

-   When adding a new social media platform:
    1.  Add new configuration fields to `config/config.go` and `.env.example`.
    2.  Create a new `platform.go` file in the `posting/` package.
    3.  Implement the `PostToPlatform` function, including actual API calls.
    4.  Update `cmd/process.go` to call the new posting function, passing necessary config.
-   When modifying GitHub interaction (e.g., how new posts are detected):
    -   Changes will primarily be in `github/git.go`.
    -   Ensure `cmd/process.go` correctly uses the output.
-   **Persistence**: The `lastProcessedCommitSHA` in `cmd/process.go` is a critical piece of state that needs to be persisted across application runs. This is a key item for future implementation. Consider simple file-based storage or a lightweight DB.

## Testing

-   (No automated tests are currently in place - this is a future enhancement)
-   When adding features or fixing bugs, consider how they would be tested.

Remember to update this `AGENTS.md` if significant architectural changes or new conventions are introduced.
