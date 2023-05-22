<p align="center">
  <img src="https://snyk.io/style/asset/logo/snyk-print.svg" />
</p>

# Snyk IaC Rules CLI Extension

## Overview

This repository contains an extension to the Snyk CLI that provides workflows to
author and manage custom rules for Snyk IaC.

## Usage

This repository produces a standalone binary for debugging purposes. This
extension is also built into the [Snyk CLI](https://github.com/snyk/cli).
Outside of debugging and development, we advise to use the Snyk CLI instead of
the standalone binary.

## Workflows

- `snyk iac rules push`
  - Builds and pushes a custom rules project to the Snyk API
  - Can also be used to delete a custom rules project from the Snyk API
- `snyk iac rules init`
  - Prompts to initialize a custom rules project, relation, rule, or spec
- `snyk iac test`
  - Tests all rules in the project against their specs
  - Also used to generate the expected output for specs
