# Bumper

> This is my version bumper. There are many like it, but this one is mine. 

## Getting Started

```bash
git clone git@github.com:drewsilcock/bumper.git
cd bumper
go build
sudo cp ./bumper /usr/local/bin/bumper

cd ~/my-project
bumper
```

## Assumptions

This is made with my personal workflow in mind, so we make certain assumptions:

- The readme is called `README.md` and contains as the first line `# {Project Name}`.
- The changelog is called `CHANGELOG.md` and contains a list of versions in the format `## v{Version} - {Date}` with the unreleased changes in a section at the top called either `## Unreleased` or `## Development`.
- Git flow is being with the development branch called `dev` and the main branch called `main`.
- Tags are added to the main branch but the tagged commits are merged into dev so that they are accessible on the dev branch.

## Configuration

The bump type is input via prompt from the user. I might add CLI arg-based input in the future.

If your remote is a GitLab repo, bumper will try to create a GitLab release which will look for an environment variable called `GITLAB_API_KEY` containing an access token with the `api` permission (no other permissions are needed).