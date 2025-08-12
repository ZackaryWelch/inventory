module.exports = {
  plugins: {
    // Tailwind v4 PostCSS plugin - handles most CSS transforms internally
    '@tailwindcss/postcss': {},
    
    // Optimized autoprefixer for modern browsers
    autoprefixer: {
      // Target modern browsers (Safari 16.4+, Chrome 111+, Firefox 128+)
      overrideBrowserslist: [
        'Safari >= 16.4',
        'Chrome >= 111', 
        'Firefox >= 128',
        'Edge >= 111'
      ],
      // Remove unnecessary prefixes for modern features
      remove: true,
      // Enable modern CSS features
      supports: true,
      // Use flexbox spec (modern only, skip 2009 spec)
      flexbox: 'no-2009'
    }
  }
};
