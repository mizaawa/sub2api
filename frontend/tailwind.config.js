/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,js,ts,jsx,tsx}'],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        // 主色调 - Teal/Cyan 青色系
        primary: {
          50: '#effcfb',
          100: '#d8f7f4',
          200: '#b5eee9',
          300: '#82dfd8',
          400: '#39c5bb',
          500: '#168f87',
          600: '#177f79',
          700: '#146b66',
          800: '#145854',
          900: '#134a47',
          950: '#082f2d'
        },
        // 辅助色 - 深蓝灰
        accent: {
          50: '#fff1f7',
          100: '#ffe3ef',
          200: '#ffc6df',
          300: '#ff98c5',
          400: '#f861a3',
          500: '#e12885',
          600: '#c31d6f',
          700: '#a3155b',
          800: '#88154d',
          900: '#721643',
          950: '#460522'
        },
        // 深色模式背景
        dark: {
          50: '#edf8f7',
          100: '#d8efec',
          200: '#b5d7d2',
          300: '#91b9b5',
          400: '#6f9995',
          500: '#557a77',
          600: '#426260',
          700: '#304b4b',
          800: '#253637',
          900: '#1b2728',
          950: '#131c1d'
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
        'gradient-primary': '#39c5bb',
        'gradient-dark': '#182223',
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
