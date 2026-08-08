/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,js,ts,jsx,tsx}'],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        // 主色调 - Teal/Cyan 青色系
        primary: {
          50: '#fff8f3',
          100: '#fce9de',
          200: '#f7cfbc',
          300: '#efa98d',
          400: '#e38361',
          500: '#d97757',
          600: '#c45f42',
          700: '#a54a33',
          800: '#873e2f',
          900: '#70372c',
          950: '#3c1b14'
        },
        // 辅助色 - 深蓝灰
        accent: {
          50: '#f8f9fa',
          100: '#f1f3f4',
          200: '#e3e6e8',
          300: '#c9cdd1',
          400: '#9da3a8',
          500: '#73787d',
          600: '#5a5f63',
          700: '#44484b',
          800: '#303336',
          900: '#202124',
          950: '#131415'
        },
        // 深色模式背景
        dark: {
          50: '#f7f6f2',
          100: '#eeece6',
          200: '#dedbd2',
          300: '#c7c3ba',
          400: '#9e9a92',
          500: '#76736d',
          600: '#5c5954',
          700: '#45433f',
          800: '#2d2c29',
          900: '#20201e',
          950: '#151514'
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
        glass: '0 1px 2px rgba(73, 45, 36, 0.08)',
        'glass-sm': '0 1px 2px rgba(73, 45, 36, 0.06)',
        glow: '0 0 0 transparent',
        'glow-lg': '0 0 0 transparent',
        card: '0 1px 2px rgba(73, 45, 36, 0.08)',
        'card-hover': '0 4px 12px rgba(73, 45, 36, 0.12)',
        'inner-glow': 'inset 0 1px 0 rgba(255, 255, 255, 0.1)'
      },
      backgroundImage: {
        'gradient-radial': 'radial-gradient(var(--tw-gradient-stops))',
        'gradient-primary': '#d9654c',
        'gradient-dark': '#302623',
        'gradient-glass':
          'linear-gradient(135deg, rgba(255,255,255,0.1) 0%, rgba(255,255,255,0.05) 100%)',
        'mesh-gradient':
          'radial-gradient(at 40% 20%, rgba(20, 184, 166, 0.12) 0px, transparent 50%), radial-gradient(at 80% 0%, rgba(6, 182, 212, 0.08) 0px, transparent 50%), radial-gradient(at 0% 50%, rgba(20, 184, 166, 0.08) 0px, transparent 50%)'
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
          '0%': { boxShadow: '0 0 20px rgba(20, 184, 166, 0.25)' },
          '100%': { boxShadow: '0 0 30px rgba(20, 184, 166, 0.4)' }
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
