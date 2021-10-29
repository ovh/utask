module.exports = {
  env: {
    browser: true,
    es2021: true,
  },
  extends: [
    'airbnb-base',
  ],
  parser: '@typescript-eslint/parser',
  parserOptions: {
    ecmaVersion: 12,
    sourceType: 'module',
  },
  plugins: [
    '@typescript-eslint',
  ],
  rules: {
    "import/prefer-default-export": "off",
    "import/no-unresolved": "off",
    "import/prefer-default-export": "off"
  },
  "object-curly-newline": ["error", {
    "ObjectExpression": "always",
    "ObjectPattern": { "multiline": true },
    "ImportDeclaration": "never",
    "ExportDeclaration": { "multiline": true, "minProperties": 3 }
  }],
  "ObjectExpression": { "multiline": true, "minProperties": 1 }
};
