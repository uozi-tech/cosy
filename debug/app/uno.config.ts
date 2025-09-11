import {
  defineConfig,
  presetAttributify,
  presetIcons,
  presetUno,
  transformerDirectives,
  transformerVariantGroup,
} from 'unocss'

export default defineConfig({
  presets: [
    presetUno(),
    presetAttributify(),
    presetIcons({
      collections: {
        'ant-design': () => import('@iconify-json/ant-design/icons.json').then(i => i.default),
      },
    }),
  ],
  transformers: [
    transformerDirectives(),
    transformerVariantGroup(),
  ],
  theme: {
    colors: {
      primary: {
        50: '#eff6ff',
        100: '#dbeafe',
        500: '#3b82f6',
        600: '#2563eb',
        700: '#1d4ed8',
        900: '#1e3a8a',
      },
      success: {
        100: '#d1fae5',
        500: '#10b981',
        600: '#059669',
        800: '#065f46',
      },
      warning: {
        100: '#fef3c7',
        500: '#f59e0b',
        600: '#d97706',
        800: '#92400e',
      },
      danger: {
        100: '#fee2e2',
        500: '#ef4444',
        600: '#dc2626',
        800: '#991b1b',
      },
    },
  },
  shortcuts: {
    'status-badge': 'px-2 py-1 text-xs font-medium rounded-full',
    'card': 'bg-white rounded-lg shadow-sm p-6 border border-gray-200',
    'btn': 'px-4 py-2 rounded font-medium transition-colors',
    'btn-primary': 'btn bg-primary-500 text-white hover:bg-primary-600',
    'btn-secondary': 'btn bg-gray-100 text-gray-700 hover:bg-gray-200',
    'nav-link': 'px-3 py-2 rounded-md text-sm font-medium transition-colors',
    'nav-link-active': 'nav-link bg-primary-100 text-primary-700',
    'nav-link-inactive': 'nav-link text-gray-500 hover:text-gray-700 hover:bg-gray-50',
  },
})
