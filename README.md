# Social Poster

Social Poster is a Go application designed to monitor a private GitHub repository for new posts (e.g., Markdown files, specific commit patterns). When new content is detected, it automatically shares this content to various configured social media platforms like Telegram, Twitter, Mastodon, Dev.to, and LinkedIn.

## Features (Planned & Implemented)

-   Monitors a private GitHub repository for new content.
-   Loads credentials securely from a `.env` file.
-   Posts to multiple social media platforms (currently placeholder implementations):
    -   Telegram
    -   Twitter
    -   Mastodon
    -   Dev.to
    -   LinkedIn
-   Dockerized for easy deployment and consistent environment.

## Getting Started

### Prerequisites

-   Docker and Docker Compose installed.
-   `git` command-line tool (for local development or if not using Docker for everything).
-   A private GitHub repository containing your posts.
-   API keys and tokens for the social media platforms you want to use.

### Configuration

1.  **Copy the Example Environment File**:
    ```bash
    cp .env.example .env
    ```

2.  **Edit `.env`**:
    Fill in the required credentials and configuration details in the `.env` file.
    Specifically, ensure the following are set:
    -   `GITHUB_USERNAME`
    -   `GITHUB_TOKEN` (Personal Access Token with `repo` scope)
    -   `POSTS_REPO_URL` (e.g., `https://github.com/YOUR_USERNAME/my_posts.git`)
    -   `POSTS_REPO_PATH=/app/my_posts_data` (This path is used by the Docker Compose volume setup to persist the cloned repository data within the container's filesystem.)
    -   Tokens and keys for each social media platform you intend to use.
    -   For Telegram, you'll also need `TELEGRAM_CHAT_ID`. (This was missing from the previous .env.example, I should add it).

### Running with Docker Compose

1.  **Build and Run the Application**:
    ```bash
    docker-compose up --build
    ```
    This command will build the Docker image (if it's the first time or if changes were made) and start the `social_poster` service. The application will then execute its `process` command by default.

2.  **To run a specific command (if different from default)**:
    You can modify the `command` directive in `docker-compose.yml` or override it:
    ```bash
    docker-compose run --rm social_poster process --some-flag
    ```
    (Currently, `process` is the main command and takes no flags).

### Development

(Details about local Go development can be added here if needed, e.g., `go run main.go process`)

## CI/CD

This project uses GitHub Actions for continuous integration and deployment.

-   **Build Go Binaries**: Automatically compiles the Go application for Linux (ARM64) and macOS (ARM64) on every push to `main`, tag push (`v*.*.*`), or pull request to `main`. The compiled binaries are available as downloadable artifacts from the workflow run.
    [![Build Go Binaries](https://github.com/MY_USERNAME/social_poster/actions/workflows/build-binaries.yml/badge.svg)](https://github.com/MY_USERNAME/social_poster/actions/workflows/build-binaries.yml)

-   **Build and Push Docker Image**: Automatically builds a Docker image and pushes it to the GitHub Container Registry (GHCR) on every push to `main` or tag push (`v*.*.*`). Images are tagged with the Git SHA, branch name, version tag, and `latest` (for pushes to `main`).
    [![Build and Push Docker Image](https://github.com/MY_USERNAME/social_poster/actions/workflows/build-docker.yml/badge.svg)](https://github.com/MY_USERNAME/social_poster/actions/workflows/build-docker.yml)

    The Docker image can be found at `ghcr.io/MY_USERNAME/social_poster`. (Replace `MY_USERNAME` with your actual GitHub username/organization).

## Project Structure

-   `main.go`: Entry point of the application.
-   `cmd/`: Cobra CLI command definitions.
    -   `root.go`: Root command setup.
    -   `process.go`: The main `process` command for fetching and posting.
-   `config/`: Configuration loading (from `.env`).
-   `github/`: GitHub interaction logic (cloning, fetching posts).
-   `posting/`: Social media posting logic (currently placeholders).
-   `Dockerfile`: Defines the Docker image.
-   `docker-compose.yml`: For orchestrating the Docker container.
-   `.env.example`: Template for environment variables.

## Security Notes

-   **Never commit your `.env` file** or any files containing live credentials to version control. The `.gitignore` file should already list `.env`.
-   The GitHub token is used for cloning the private repository. Ensure it has the minimum necessary permissions (usually `repo` scope).
-   Social media API keys and tokens should be treated with the same level of care.

## Future Enhancements (TODO)

-   Implement actual API calls for each social media platform in the `posting/` package.
-   Develop a robust mechanism for detecting "new" posts in the `github/git.go` (e.g., based on file changes in new commits, specific naming conventions for post files, parsing commit messages).
-   Implement persistent storage for `lastProcessedCommitSHA` to correctly track processed posts across application restarts.
-   Add more sophisticated error handling and retry mechanisms for posting.
-   Implement content formatting and adaptation for different social media platforms.
-   Add unit and integration tests.
-   Consider a more secure method for handling the GitHub token during git operations if shelling out to `git clone/pull` remains the chosen method (e.g., git credential helper, or switch to a Go git library).

This project is under development.
