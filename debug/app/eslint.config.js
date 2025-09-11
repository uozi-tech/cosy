import antfu from '@antfu/eslint-config'

export default antfu({
  stylistic: true,
  typescript: true,
  vue: true,
  formatters: true,
  rules: {
    'vue/component-name-in-template-casing': ['error', 'PascalCase', {
      registeredComponentsOnly: false,
      ignores: [],
    }],
    'vue/component-tags-order': ['error', {
      order: ['script', 'template', 'style'],
    }],
  },
})
