# OpenGov UI Style Guide

**Version 1.0** | Last Updated: 2025-11-13

---

## Table of Contents
1. [Brand Identity](#brand-identity)
2. [Color System](#color-system)
3. [Typography](#typography)
4. [Spacing & Layout](#spacing--layout)
5. [Components](#components)
6. [Icons & Imagery](#icons--imagery)
7. [Interactive States](#interactive-states)
8. [Accessibility](#accessibility)
9. [Animation & Motion](#animation--motion)

---

## Brand Identity

### Overview
OpenGov is a government transparency platform that makes federal register information accessible, digestible, and shareable. Our design language should convey:

- **Trust**: Professional, reliable, authoritative
- **Clarity**: Clean, readable, well-organized
- **Accessibility**: Inclusive, easy to understand
- **Modern**: Contemporary without being trendy

### Voice & Tone
- **Professional yet approachable**: Not stuffy or bureaucratic
- **Clear and concise**: Avoid jargon, use plain language
- **Informative**: Educational without being condescending
- **Neutral**: Unbiased presentation of government information

---

## Color System

### Primary Colors

```css
/* Primary Blue - Main brand color */
--color-primary-50:  #eff6ff;
--color-primary-100: #dbeafe;
--color-primary-200: #bfdbfe;
--color-primary-300: #93c5fd;
--color-primary-400: #60a5fa;
--color-primary-500: #3b82f6;  /* Main */
--color-primary-600: #2563eb;  /* Hover */
--color-primary-700: #1d4ed8;  /* Active */
--color-primary-800: #1e40af;
--color-primary-900: #1e3a8a;
```

**Usage:**
- Primary actions (CTAs, buttons, links)
- Interactive elements
- Focus states
- Active navigation items

### Secondary Colors

```css
/* Neutral Gray - Text and backgrounds */
--color-neutral-50:  #f9fafb;
--color-neutral-100: #f3f4f6;
--color-neutral-200: #e5e7eb;
--color-neutral-300: #d1d5db;
--color-neutral-400: #9ca3af;
--color-neutral-500: #6b7280;
--color-neutral-600: #4b5563;
--color-neutral-700: #374151;
--color-neutral-800: #1f2937;
--color-neutral-900: #111827;
```

**Usage:**
- Body text: `neutral-700`
- Headings: `neutral-900`
- Borders: `neutral-200`
- Background: `neutral-50`
- Disabled states: `neutral-300`

### Semantic Colors

```css
/* Success - Green */
--color-success-50:  #f0fdf4;
--color-success-500: #22c55e;
--color-success-600: #16a34a;
--color-success-700: #15803d;

/* Warning - Amber */
--color-warning-50:  #fffbeb;
--color-warning-500: #f59e0b;
--color-warning-600: #d97706;
--color-warning-700: #b45309;

/* Error - Red */
--color-error-50:  #fef2f2;
--color-error-500: #ef4444;
--color-error-600: #dc2626;
--color-error-700: #b91c1c;

/* Info - Cyan */
--color-info-50:  #ecfeff;
--color-info-500: #06b6d4;
--color-info-600: #0891b2;
--color-info-700: #0e7490;
```

**Usage:**
- Success states, confirmations
- Warnings and cautions
- Error messages and validation
- Informational alerts

### Government Category Colors

Use these accent colors to categorize different types of federal register documents:

```css
--category-executive:    #8b5cf6;  /* Purple - Executive Orders */
--category-rules:        #0891b2;  /* Cyan - Rules & Regulations */
--category-notices:      #f59e0b;  /* Amber - Public Notices */
--category-presidential: #dc2626;  /* Red - Presidential Documents */
```

---

## Typography

### Type Scale

```css
/* Font Families */
--font-sans: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', system-ui, sans-serif;
--font-mono: 'JetBrains Mono', 'Fira Code', 'Courier New', monospace;

/* Font Sizes */
--text-xs:   0.75rem;   /* 12px */
--text-sm:   0.875rem;  /* 14px */
--text-base: 1rem;      /* 16px */
--text-lg:   1.125rem;  /* 18px */
--text-xl:   1.25rem;   /* 20px */
--text-2xl:  1.5rem;    /* 24px */
--text-3xl:  1.875rem;  /* 30px */
--text-4xl:  2.25rem;   /* 36px */
--text-5xl:  3rem;      /* 48px */
--text-6xl:  3.75rem;   /* 60px */

/* Line Heights */
--leading-none:    1;
--leading-tight:   1.25;
--leading-snug:    1.375;
--leading-normal:  1.5;
--leading-relaxed: 1.625;
--leading-loose:   2;

/* Font Weights */
--font-normal:    400;
--font-medium:    500;
--font-semibold:  600;
--font-bold:      700;
```

### Heading Styles

```css
/* H1 - Page Titles */
h1 {
  font-size: var(--text-4xl);
  font-weight: var(--font-bold);
  line-height: var(--leading-tight);
  color: var(--color-neutral-900);
  letter-spacing: -0.025em;
}

/* H2 - Section Headers */
h2 {
  font-size: var(--text-3xl);
  font-weight: var(--font-bold);
  line-height: var(--leading-tight);
  color: var(--color-neutral-900);
  letter-spacing: -0.025em;
}

/* H3 - Subsection Headers */
h3 {
  font-size: var(--text-2xl);
  font-weight: var(--font-semibold);
  line-height: var(--leading-snug);
  color: var(--color-neutral-800);
}

/* H4 - Card Titles */
h4 {
  font-size: var(--text-xl);
  font-weight: var(--font-semibold);
  line-height: var(--leading-snug);
  color: var(--color-neutral-800);
}

/* H5 - Small Headers */
h5 {
  font-size: var(--text-lg);
  font-weight: var(--font-medium);
  line-height: var(--leading-normal);
  color: var(--color-neutral-700);
}

/* H6 - Labels */
h6 {
  font-size: var(--text-base);
  font-weight: var(--font-medium);
  line-height: var(--leading-normal);
  color: var(--color-neutral-700);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}
```

### Body Text

```css
/* Default Body */
body, p {
  font-size: var(--text-base);
  font-weight: var(--font-normal);
  line-height: var(--leading-relaxed);
  color: var(--color-neutral-700);
}

/* Large Body - Article intros */
.text-large {
  font-size: var(--text-lg);
  line-height: var(--leading-relaxed);
}

/* Small Body - Metadata, captions */
.text-small {
  font-size: var(--text-sm);
  line-height: var(--leading-normal);
  color: var(--color-neutral-600);
}

/* Extra Small - Labels, tags */
.text-xs {
  font-size: var(--text-xs);
  line-height: var(--leading-normal);
  color: var(--color-neutral-500);
  font-weight: var(--font-medium);
}
```

---

## Spacing & Layout

### Spacing Scale

Use consistent spacing based on a 4px base unit:

```css
--space-0:  0;
--space-1:  0.25rem;  /* 4px */
--space-2:  0.5rem;   /* 8px */
--space-3:  0.75rem;  /* 12px */
--space-4:  1rem;     /* 16px */
--space-5:  1.25rem;  /* 20px */
--space-6:  1.5rem;   /* 24px */
--space-8:  2rem;     /* 32px */
--space-10: 2.5rem;   /* 40px */
--space-12: 3rem;     /* 48px */
--space-16: 4rem;     /* 64px */
--space-20: 5rem;     /* 80px */
--space-24: 6rem;     /* 96px */
```

### Container Widths

```css
--container-sm:  640px;   /* Small devices */
--container-md:  768px;   /* Tablets */
--container-lg:  1024px;  /* Desktops */
--container-xl:  1280px;  /* Large desktops */
--container-2xl: 1536px;  /* Extra large */
```

### Grid System

Use a 12-column grid with consistent gutters:

```css
.grid {
  display: grid;
  grid-template-columns: repeat(12, 1fr);
  gap: var(--space-6);
}

/* Responsive breakpoints */
@media (max-width: 768px) {
  .grid {
    gap: var(--space-4);
  }
}
```

### Layout Patterns

#### Page Container
```jsx
<div className="container mx-auto px-4 md:px-6 lg:px-8 max-w-7xl">
  {/* Content */}
</div>
```

#### Section Spacing
- Between major sections: `var(--space-16)` to `var(--space-24)`
- Between subsections: `var(--space-8)` to `var(--space-12)`
- Between related elements: `var(--space-4)` to `var(--space-6)`

---

## Components

### Buttons

#### Primary Button
```jsx
<button className="
  px-6 py-3
  bg-primary-600 hover:bg-primary-700 active:bg-primary-800
  text-white font-medium text-base
  rounded-lg
  transition-colors duration-150
  focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2
  disabled:bg-neutral-300 disabled:cursor-not-allowed
">
  Button Text
</button>
```

**Sizes:**
- Small: `px-4 py-2 text-sm`
- Medium (default): `px-6 py-3 text-base`
- Large: `px-8 py-4 text-lg`

#### Secondary Button
```jsx
<button className="
  px-6 py-3
  bg-white hover:bg-neutral-50 active:bg-neutral-100
  text-primary-600 font-medium text-base
  border border-neutral-300
  rounded-lg
  transition-colors duration-150
  focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2
">
  Button Text
</button>
```

#### Ghost Button
```jsx
<button className="
  px-6 py-3
  bg-transparent hover:bg-neutral-100 active:bg-neutral-200
  text-neutral-700 font-medium text-base
  rounded-lg
  transition-colors duration-150
  focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2
">
  Button Text
</button>
```

### Cards

#### Article Card
```jsx
<article className="
  bg-white
  rounded-xl
  shadow-sm hover:shadow-md
  border border-neutral-200
  overflow-hidden
  transition-shadow duration-200
">
  <div className="p-6">
    {/* Category badge */}
    <span className="inline-flex items-center px-3 py-1 rounded-full text-xs font-medium bg-primary-100 text-primary-700">
      Executive Order
    </span>

    {/* Title */}
    <h3 className="mt-4 text-xl font-semibold text-neutral-900 line-clamp-2">
      Article Title Goes Here
    </h3>

    {/* Excerpt */}
    <p className="mt-2 text-base text-neutral-600 line-clamp-3">
      Brief summary of the article content...
    </p>

    {/* Metadata */}
    <div className="mt-4 flex items-center gap-4 text-sm text-neutral-500">
      <span>Jan 13, 2025</span>
      <span>•</span>
      <span>5 min read</span>
    </div>
  </div>
</article>
```

### Forms

#### Input Field
```jsx
<div className="space-y-1">
  <label className="block text-sm font-medium text-neutral-700">
    Label
  </label>
  <input
    type="text"
    className="
      w-full px-4 py-2
      text-base text-neutral-900
      bg-white
      border border-neutral-300 rounded-lg
      placeholder:text-neutral-400
      focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent
      disabled:bg-neutral-50 disabled:text-neutral-500
    "
    placeholder="Enter text..."
  />
  <p className="text-sm text-neutral-500">Helper text</p>
</div>
```

#### Error State
```jsx
<input
  className="
    border-error-500
    focus:ring-error-500
  "
/>
<p className="text-sm text-error-600">Error message</p>
```

### Badges

```jsx
/* Category Badge */
<span className="inline-flex items-center px-3 py-1 rounded-full text-xs font-medium bg-primary-100 text-primary-700">
  Badge Text
</span>

/* Status Badge - Success */
<span className="inline-flex items-center px-3 py-1 rounded-full text-xs font-medium bg-success-100 text-success-700">
  Active
</span>

/* Status Badge - Warning */
<span className="inline-flex items-center px-3 py-1 rounded-full text-xs font-medium bg-warning-100 text-warning-700">
  Pending
</span>
```

### Navigation

#### Header Navigation
```jsx
<header className="sticky top-0 z-50 bg-white border-b border-neutral-200">
  <nav className="container mx-auto px-4 md:px-6">
    <div className="flex items-center justify-between h-16">
      {/* Logo */}
      <div className="flex items-center gap-2">
        <img src="/logo.svg" alt="OpenGov" className="h-8 w-8" />
        <span className="text-xl font-bold text-neutral-900">OpenGov</span>
      </div>

      {/* Nav Links */}
      <div className="hidden md:flex items-center gap-8">
        <a href="/feed" className="text-base font-medium text-neutral-700 hover:text-primary-600 transition-colors">
          Feed
        </a>
        <a href="/about" className="text-base font-medium text-neutral-700 hover:text-primary-600 transition-colors">
          About
        </a>
      </div>

      {/* Actions */}
      <button className="px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700">
        Sign In
      </button>
    </div>
  </nav>
</header>
```

### Loading States

#### Skeleton Loader
```jsx
<div className="animate-pulse">
  <div className="h-4 bg-neutral-200 rounded w-3/4 mb-4"></div>
  <div className="h-4 bg-neutral-200 rounded w-1/2 mb-4"></div>
  <div className="h-4 bg-neutral-200 rounded w-5/6"></div>
</div>
```

#### Spinner
```jsx
<div className="inline-block h-8 w-8 animate-spin rounded-full border-4 border-solid border-primary-600 border-r-transparent"></div>
```

---

## Icons & Imagery

### Icon System

Use **Heroicons** (outline style) for consistency:

```jsx
import { DocumentTextIcon, ShareIcon, UserIcon } from '@heroicons/react/24/outline';

<DocumentTextIcon className="h-5 w-5 text-neutral-700" />
```

**Icon Sizes:**
- Extra Small: `h-4 w-4` (16px)
- Small: `h-5 w-5` (20px)
- Medium: `h-6 w-6` (24px)
- Large: `h-8 w-8` (32px)
- Extra Large: `h-12 w-12` (48px)

### Image Guidelines

#### Article Thumbnails
- Aspect ratio: 16:9
- Minimum resolution: 1200x675px
- File format: WebP with JPG fallback
- Max file size: 200KB

#### Placeholder Images
Use a neutral gray background with icon:
```jsx
<div className="aspect-video bg-neutral-100 flex items-center justify-center">
  <DocumentTextIcon className="h-12 w-12 text-neutral-400" />
</div>
```

---

## Interactive States

### State Specifications

```css
/* Default State */
.interactive {
  transition: all 150ms cubic-bezier(0.4, 0, 0.2, 1);
}

/* Hover State */
.interactive:hover {
  /* Slight brightness increase for backgrounds */
  /* Color shift for text */
  transform: translateY(-1px); /* Subtle lift for cards */
}

/* Active/Pressed State */
.interactive:active {
  transform: translateY(0); /* Return to normal */
  /* Darker/more saturated color */
}

/* Focus State */
.interactive:focus {
  outline: none;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.5); /* Primary color with opacity */
}

/* Disabled State */
.interactive:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  pointer-events: none;
}
```

### Link Styles

```css
/* Inline Text Link */
a {
  color: var(--color-primary-600);
  text-decoration: none;
  border-bottom: 1px solid transparent;
  transition: border-color 150ms;
}

a:hover {
  border-bottom-color: var(--color-primary-600);
}

a:active {
  color: var(--color-primary-700);
}
```

---

## Accessibility

### WCAG 2.1 AA Compliance

#### Color Contrast
- Normal text (< 18px): Minimum 4.5:1
- Large text (≥ 18px or 14px bold): Minimum 3:1
- UI components and graphics: Minimum 3:1

**Approved Color Combinations:**
- `neutral-900` on `white` ✓ (16.1:1)
- `neutral-700` on `white` ✓ (7.7:1)
- `white` on `primary-600` ✓ (4.8:1)
- `white` on `primary-700` ✓ (6.3:1)

#### Focus Indicators
All interactive elements must have visible focus indicators:
```css
:focus-visible {
  outline: 2px solid var(--color-primary-500);
  outline-offset: 2px;
}
```

#### Keyboard Navigation
- All interactive elements must be keyboard accessible
- Tab order should be logical and intuitive
- Provide skip links for repetitive content
- Use semantic HTML elements

#### Screen Reader Support
```jsx
/* Hidden but accessible labels */
<span className="sr-only">Description for screen readers</span>

/* ARIA labels */
<button aria-label="Close dialog">
  <XMarkIcon className="h-6 w-6" />
</button>

/* ARIA live regions for dynamic content */
<div role="status" aria-live="polite">
  {statusMessage}
</div>
```

#### Motion & Animation
Respect user preferences:
```css
@media (prefers-reduced-motion: reduce) {
  * {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
  }
}
```

---

## Animation & Motion

### Timing Functions

```css
/* Easing curves */
--ease-in:      cubic-bezier(0.4, 0, 1, 1);
--ease-out:     cubic-bezier(0, 0, 0.2, 1);
--ease-in-out:  cubic-bezier(0.4, 0, 0.2, 1);
--ease-spring:  cubic-bezier(0.68, -0.55, 0.265, 1.55);
```

### Duration Scale

```css
--duration-fast:   150ms;  /* Micro-interactions */
--duration-base:   200ms;  /* Hover states */
--duration-slow:   300ms;  /* Modals, dropdowns */
--duration-slower: 500ms;  /* Page transitions */
```

### Animation Patterns

#### Fade In
```css
@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

.fade-in {
  animation: fadeIn var(--duration-base) var(--ease-out);
}
```

#### Slide Up
```css
@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.slide-up {
  animation: slideUp var(--duration-slow) var(--ease-out);
}
```

#### Stagger Children
```jsx
/* Stagger animation for list items */
{items.map((item, index) => (
  <div
    key={item.id}
    className="animate-fade-in"
    style={{ animationDelay: `${index * 50}ms` }}
  >
    {item.content}
  </div>
))}
```

---

## Best Practices

### Component Development

1. **Use shadcn/ui as the foundation**: Customize shadcn components rather than building from scratch
2. **Composition over customization**: Build complex components from simpler ones
3. **Consistent spacing**: Always use spacing scale variables
4. **Responsive by default**: Mobile-first approach for all components
5. **Type safety**: Use TypeScript interfaces for all component props

### Performance

1. **Optimize images**: Use WebP format, lazy loading, and responsive images
2. **Minimize re-renders**: Use React.memo() and useMemo() appropriately
3. **Code splitting**: Use dynamic imports for routes and large components
4. **Bundle size**: Monitor and optimize dependencies

### Maintenance

1. **Document patterns**: Update this guide when creating new patterns
2. **Design tokens**: Use CSS variables for all design values
3. **Version control**: Track design decisions in this document
4. **Component library**: Maintain a Storybook for component documentation

---

## Resources

### Design Tools
- [Figma](https://figma.com) - UI design and prototyping
- [Excalidraw](https://excalidraw.com) - Wireframing and diagrams

### Development Tools
- [shadcn/ui](https://ui.shadcn.com/) - Component library
- [Tailwind CSS](https://tailwindcss.com/) - Utility-first CSS
- [Heroicons](https://heroicons.com/) - Icon system
- [Inter Font](https://rsms.me/inter/) - Primary typeface

### Accessibility
- [WCAG 2.1 Guidelines](https://www.w3.org/WAI/WCAG21/quickref/)
- [WebAIM Color Contrast Checker](https://webaim.org/resources/contrastchecker/)
- [a11y Project Checklist](https://www.a11yproject.com/checklist/)

---

**Document Maintenance**: This style guide should be updated whenever new design patterns are established or existing patterns are modified. All team members are responsible for keeping this document current and accurate.
