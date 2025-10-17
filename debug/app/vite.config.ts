import process from 'node:process'
import { fileURLToPath, URL } from 'node:url'
import vue from '@vitejs/plugin-vue'
import vueJsx from '@vitejs/plugin-vue-jsx'
import UnoCSS from 'unocss/vite'
import AutoImport from 'unplugin-auto-import/vite'
import { AntDesignVueResolver } from 'unplugin-vue-components/resolvers'
import Components from 'unplugin-vue-components/vite'
import { defineConfig, loadEnv } from 'vite'

// https://vitejs.dev/config/
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')

  return {
    base: './',
    plugins: [
      vue(),
      vueJsx(),
      UnoCSS(),
      AutoImport({
        imports: [
          'vue',
          'vue-router',
          'pinia',
          '@vueuse/core',
        ],
        vueTemplate: true,
        eslintrc: {
          enabled: true,
          filepath: '.eslint-auto-import.mjs',
        },
      }),
      Components({
        resolvers: [
          AntDesignVueResolver({ importStyle: false }),
        ],
        directoryAsNamespace: true,
      }),
    ],
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url)),
      },
      extensions: [
        '.mjs',
        '.js',
        '.ts',
        '.jsx',
        '.tsx',
        '.json',
        '.vue',
      ],
    },
    css: {
      preprocessorOptions: {
        less: {
          modifyVars: {
            'border-radius-base': '5px',
          },
          javascriptEnabled: true,
        },
      },
    },
    build: {
      outDir: 'dist',
      assetsDir: 'assets',
      chunkSizeWarningLimit: 1500,
      rollupOptions: {
        output: {
          manualChunks: {
            'ant-design-vue': ['ant-design-vue'],
            'vue-vendor': ['vue', 'vue-router', 'pinia'],
          },
        },
      },
    },
    server: {
      port: Number.parseInt(env.VITE_PORT) || 3000,
      proxy: {
        '/api/debug': {
          target: env.VITE_PROXY_TARGET || 'http://localhost:9001',
          changeOrigin: true,
          secure: false,
          ws: true,
          rewrite: path => path.replace(/^\/api/, ''),
        },
      },
    },
  }
})
