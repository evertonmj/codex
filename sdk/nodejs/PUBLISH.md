# Publishing to npm

To publish the CodexDB Node.js SDK to npm:

1. Make sure you have an npm account: https://www.npmjs.com/signup
2. Login to npm in your terminal:
   ```sh
   npm login
   ```
3. In the sdk/nodejs directory, publish the package:
   ```sh
   npm publish
   ```

If you get a name conflict error, choose a unique package name in package.json (e.g., codexdb-sdk-everton) and try again.

## Versioning
- Update the version field in package.json before publishing a new release.

## Automated Publishing (optional)
You can automate npm publishing using GitHub Actions or another CI/CD tool. Let me know if you want a workflow example!
