# UI Style Guide

## Color System

### Semantic Colors
- **Primary**: Dark gray (900) - Main brand color, CTAs
- **Secondary**: Light gray (100) - Secondary elements
- **Destructive**: Red (600) - Error states, warnings
- **Muted**: Gray (200) - Disabled, subtle text
- **Accent**: Dark gray (900) - Highlights

### Color Palette
```
Background:     #FFFFFF (white)
Foreground:     #0A0A0A (nearly black)
Border:         #E5E5E5 (light gray)
Muted:          #F5F5F5 (very light gray)
```

## Typography

### Font Stack
```
-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Oxygen, Ubuntu, 
Cantarell, "Fira Sans", "Droid Sans", "Helvetica Neue", sans-serif
```

### Scale
- **H1/H2**: 24px, font-weight 700
- **H3**: 20px, font-weight 600
- **Body**: 16px, font-weight 400
- **Small**: 14px, font-weight 400
- **Tiny**: 12px, font-weight 400

## Components

### ArticleCard
- Border: 1px solid border
- Padding: 24px
- Hover: shadow-lg
- Content:
  - Title (h2, 20px, font-bold, max 2 lines)
  - Summary (body, muted foreground, max 3 lines)
  - Footer: date (left) + source link (right)

### FeedList
- Gap: 24px between cards
- Pagination: centered controls at bottom
- Loading: 3 skeleton cards, h-32, animate-pulse

## Responsive Design

### Breakpoints
- Mobile: < 640px
- Tablet: 640px - 1024px
- Desktop: > 1024px

### Touch Targets
- Minimum 44px × 44px for interactive elements
- Padding around buttons/links

### Mobile Optimizations
- Single column layout
- Full-width cards with 16px padding
- Larger touch targets (48px)
- Simplified navigation

## Loading States
- Skeleton loaders with `animate-pulse`
- Gray background placeholder height
- Progressive reveal as content loads

## Error States
- Red background (#FEE2E2)
- Red border (#DC2626)
- Red text (#991B1B)
- Clear messaging
