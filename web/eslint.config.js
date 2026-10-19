import js from '@eslint/js'
import globals from 'globals'
import stylistic from '@stylistic/eslint-plugin'
import reactHooks from 'eslint-plugin-react-hooks'
import tseslint from 'typescript-eslint'

export default tseslint.config(
  { ignores: ['dist'] },
  js.configs.recommended,
  tseslint.configs.recommended,
  {
    languageOptions: {
      ecmaVersion: 2022,
      sourceType: 'module',
      globals: globals.browser,
    },
    plugins: { '@stylistic': stylistic },
    // House style, mechanised rather than remembered.
    rules: {
      '@stylistic/semi': ['error', 'never'],
      '@stylistic/quotes': ['error', 'single', { avoidEscape: true }],
    },
  },
  {
    files: ['**/*.{ts,tsx}'],
    extends: [reactHooks.configs.flat['recommended-latest']],
  },
)
