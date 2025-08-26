# Grob CLI 🚀

grob-cli is the official command-line interface for the Grob Framework. This tool is designed to streamline your development workflow by automating project setup, application scaffolding, and module creation.

Inspired by the developer experience of tools like the Nest.js CLI, grob-cli handles the boilerplate so you can focus on writing business logic.

## Key Features

    * Project Scaffolding: Create a new, production-ready Grob project with a single command (grob new <project-name>).

    * Application Generation: Easily add new, independent web applications to your monorepo (grob create-app <app-name>).

    * Module Creation: Generate feature modules, including controllers and services, with automatic dependency registration (grob create-module <app-name> <module-name>).

    * Automated Wiring: The CLI intelligently modifies your source code to import and register new applications and modules, ensuring everything is connected correctly.

Installation

To install the CLI, run the following command:
Bash
```
go install github.com/yuliussmayoru/grob-cli@latest
```
You can then run grob --help to see the available commands.