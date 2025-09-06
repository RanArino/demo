# Requirements

## 1. Overview

The current frontend codebase uses JavaScript files generated from protobuf definitions. This leads to warnings and duplicated type declarations. The goal is to switch to TypeScript for the generated protobuf code to improve type safety and reduce code duplication.

## 2. User Stories

- As a developer, I want the protobuf code generation to produce TypeScript files so that I can benefit from static typing.
- As a developer, I want to remove the old JavaScript protobuf files to avoid confusion and reduce the bundle size.
- As a developer, I want to update all the frontend code that uses the old JavaScript protobuf files to use the new TypeScript files, so that the application works correctly with the new generated code.

## 3. Acceptance Criteria

- The protobuf generation process must output only TypeScript files.
- All frontend code that previously used the JavaScript protobuf objects should be refactored to use the TypeScript objects.
- The application must build and run without errors after the refactoring.
- There should be no type-related warnings from the protobuf generated code.
