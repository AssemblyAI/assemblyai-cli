#!/bin/bash

go build -ldflags "-X github.com/AssemblyAI/assemblyai-cli/cmd.PH_TOKEN=$POSTHOG_API_TOKEN -X github.com/AssemblyAI/assemblyai-cli/cmd.VERSION=$VERSION -X github.com/AssemblyAI/assemblyai-cli/cmd.SENTRY_DNS=$SENTRY_DNS" -o assemblyai