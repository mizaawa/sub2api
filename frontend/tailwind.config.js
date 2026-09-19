/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,js,ts,jsx,tsx}'],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        // 主色调 - 柔和鹅黄色
        primary: {
          50: '#fffdf1',
          100: '#fff8d9',
          200: '#f8eab0',
          300: '#edd889',
          400: '#dfc45f',
          500: '#d2b455',
          600: '#b99938',
          700: '#9a7d2d',
          800: '#7b6428',
          900: '#655322',
          950: '#443818'
        },
        // 辅助色 - 莫奈式鼠尾草绿
        accent: {
          50: '#f1f4ed',
          100: '#e2e9dc',
          200: '#c7d4bd',
          300: '#a9bda0',
          400: '#8ea98a',
          500: '#78927a',
          600: '#657b67',
          700: '#526454',
          800: '#424f43',
          900: '#37413a',
          950: '#252c28'
        },
        // 兼容旧有 dark-* 工具类，但仍保持浅色画布
        dark: {
          50: '#fffdf1',
          100: '#fff8d9',
          200: '#f8eab0',
          300: '#edd889',
          400: '#dfc45f',
          500: '#d2b455',
          600: '#b99938',
          700: '#9a7d2d',
          800: '#f3eed7',
          900: '#fffdf5',
          950: '#fcf8e8'
        }
      },
      fontFamily: {
        sans: [
          'system-ui',
          '-apple-system',
          'BlinkMacSystemFont',
          'Segoe UI',
          'Roboto',
          'Helvetica Neue',
          'Arial',
          'PingFang SC',
          'Hiragino Sans GB',
          'Microsoft YaHei',
          'sans-serif'
        ],
        mono: ['ui-monospace', 'SFMono-Regular', 'Menlo', 'Monaco', 'Consolas', 'monospace']
      },
      boxShadow: {
        glass: '0 1px 2px rgba(12, 65, 62, 0.08)',
        'glass-sm': '0 1px 2px rgba(12, 65, 62, 0.06)',
        glow: '0 0 0 transparent',
        'glow-lg': '0 0 0 transparent',
        card: '0 1px 2px rgba(12, 65, 62, 0.08)',
        'card-hover': '0 4px 12px rgba(12, 65, 62, 0.12)',
        'inner-glow': 'inset 0 1px 0 rgba(255, 255, 255, 0.1)'
      },
      backgroundImage: {
        'gradient-radial': 'radial-gradient(var(--tw-gradient-stops))',
        'gradient-primary': '#d2b455',
        'gradient-dark': '#f3eed7',
        'gradient-glass':
          'linear-gradient(135deg, rgba(255,255,255,0.1) 0%, rgba(255,255,255,0.05) 100%)',
        'mesh-gradient':
          'radial-gradient(at 40% 20%, rgba(210, 180, 85, 0.16) 0px, transparent 50%), radial-gradient(at 80% 0%, rgba(142, 169, 138, 0.12) 0px, transparent 50%), radial-gradient(at 0% 50%, rgba(201, 167, 160, 0.1) 0px, transparent 50%)'
      },
      animation: {
        'fade-in': 'fadeIn 0.3s ease-out',
        'slide-up': 'slideUp 0.3s ease-out',
        'slide-down': 'slideDown 0.3s ease-out',
        'slide-in-right': 'slideInRight 0.3s ease-out',
        'scale-in': 'scaleIn 0.2s ease-out',
        'pulse-slow': 'pulse 3s cubic-bezier(0.4, 0, 0.6, 1) infinite',
        shimmer: 'shimmer 2s linear infinite',
        glow: 'glow 2s ease-in-out infinite alternate'
      },
      keyframes: {
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' }
        },
        slideUp: {
          '0%': { opacity: '0', transform: 'translateY(10px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' }
        },
        slideDown: {
          '0%': { opacity: '0', transform: 'translateY(-10px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' }
        },
        slideInRight: {
          '0%': { opacity: '0', transform: 'translateX(20px)' },
          '100%': { opacity: '1', transform: 'translateX(0)' }
        },
        scaleIn: {
          '0%': { opacity: '0', transform: 'scale(0.95)' },
          '100%': { opacity: '1', transform: 'scale(1)' }
        },
        shimmer: {
          '0%': { backgroundPosition: '-200% 0' },
          '100%': { backgroundPosition: '200% 0' }
        },
        glow: {
          '0%': { boxShadow: '0 0 20px rgba(210, 180, 85, 0.25)' },
          '100%': { boxShadow: '0 0 30px rgba(210, 180, 85, 0.4)' }
        }
      },
      backdropBlur: {
        xs: '2px'
      },
      borderRadius: {
        '4xl': '2rem',
        '5xl': '2.5rem'
      }
    }
  },
  plugins: []
}
