/** @type {import('tailwindcss').Config} */

module.exports = {
  content: ['./src/**/*.{ts,tsx}'],
  
  safelist: [
    // Icon size utilities
    'w-2', 'h-2', 'w-2.5', 'h-2.5', 'w-3', 'h-3', 'w-3.5', 'h-3.5',
    'w-4', 'h-4', 'w-4.5', 'h-4.5', 'w-5', 'h-5', 'w-6', 'h-6',
    'w-7', 'h-7', 'w-8', 'h-8', 'w-9', 'h-9', 'w-10', 'h-10',
    'w-11', 'h-11', 'w-12', 'h-12', 'w-14', 'h-14', 'w-16', 'h-16',
  ],
  
  theme: {
    extend: {
      fontSize: {
        '2xs': ['0.625rem', { lineHeight: '0.75rem' }],
      },

      fontFamily: {
        outfit: ['var(--font-outfit)', 'system-ui', 'sans-serif'],
      },

      spacing: {
        4.5: '1.125rem',
        18: '4.5rem',
      },

      colors: {
        'primary-lightest': 'var(--color-primary-lightest)',
        'primary-light': 'var(--color-primary-light)', 
        primary: 'var(--color-primary)',
        'primary-dark': 'var(--color-primary-dark)',
        accent: 'var(--color-accent)',
        'accent-dark': 'var(--color-accent-dark)',
        danger: 'var(--color-danger)',
        'danger-dark': 'var(--color-danger-dark)',
        'gray-lightest': 'var(--color-gray-lightest)',
        'gray-light': 'var(--color-gray-light)',
        gray: 'var(--color-gray)',
        'gray-dark': 'var(--color-gray-dark)',
        overlay: 'var(--color-overlay)',
        white: 'var(--color-white)',
        black: 'var(--color-black)',
      },

      borderRadius: {
        xs: '0.125rem',
        sm: '0.25rem',
        DEFAULT: '0.625rem',
        '2xl': '1rem',
        '3xl': '1.5rem',
      },

      boxShadow: {
        around: '0 0 8px 4px rgba(0, 0, 0, 0.1)',
      },

      keyframes: {
        fadeIn: {
          from: { opacity: '0' },
          to: { opacity: '1' },
        },
        fadeOut: {
          from: { opacity: '1' },
          to: { opacity: '0' },
        },
        slideInFromBottom: {
          from: { transform: 'translateY(100%)' },
          to: { transform: 'translateY(0)' },
        },
        slideOutToBottom: {
          from: { transform: 'translateY(0)' },
          to: { transform: 'translateY(100%)' },
        },
        slideInFromRight: {
          from: { transform: 'translateX(100%)' },
          to: { transform: 'translateX(0)' },
        },
        slideOutToRight: {
          from: { transform: 'translateX(0)' },
          to: { transform: 'translateX(100%)' },
        },
        scaleIn: {
          from: { transform: 'scale(0.95)', opacity: '0' },
          to: { transform: 'scale(1)', opacity: '1' },
        },
      },

      animation: {
        fadeIn: 'fadeIn 0.2s ease-out',
        fadeOut: 'fadeOut 0.15s ease-in',
        slideInFromBottom: 'slideInFromBottom 0.3s ease-out',
        slideOutToBottom: 'slideOutToBottom 0.2s ease-in',
        slideInFromRight: 'slideInFromRight 0.3s ease-out',
        slideOutToRight: 'slideOutToRight 0.2s ease-in',
        scaleIn: 'scaleIn 0.2s ease-out',
      },
    },
  },
  
  plugins: [],
};
