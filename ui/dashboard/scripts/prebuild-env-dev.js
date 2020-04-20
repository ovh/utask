#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

const ejs = require('ejs');

const files = [
  {
    directory: path.join(__dirname, '../src/environments'),
    template: 'environment.ts.template',
    file: 'environment.ts',
    defaultValues: {
      PREFIX_API_BASE_URL: '/api/',
      PREFIX_LOCALSTORAGE: 'utask-',
      SENTRY_DSN: '',
    }
  }
];

files.forEach(file => {
  const environmentTemplate = fs.readFileSync(
    path.join(file.directory, file.template),
    { encoding: 'utf-8' }
  );
  const output = ejs.render(environmentTemplate, Object.assign({}, file.defaultValues, process.env));
  fs.writeFileSync(path.join(file.directory, file.file), output);
});

process.exit(0);
