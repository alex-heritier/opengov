# OpenGov UI Style Guide

metadata:
  name: OpenGov UI Style Guide
  version: 2.0.0
  audience: Americans seeking government transparency
  philosophy: Trust through clarity. Professional yet approachable. Information-first.
  lastUpdated: 2025-11-13

changelog:
  - version: 2.0.0
    date: 2025-11-13
    changes:
      - Restructured to formalized design token system
      - Added component specifications with examples
      - Consolidated accessibility requirements with test criteria
      - Added category colors for federal documents

theme:
  mode: light
  coreRules:
    - Trust and clarity above all
    - Clean, readable layouts
    - Professional, authoritative presentation
    - Accessible by design
    - Modern without trends
    - Semantic color usage

tokens:
  color:
    primary: '{colors.light.interactive.primary}'
    text: '{colors.light.text.primary}'
    border: '{colors.light.border.default}'
    bg: '{colors.light.background.primary}'
  space:
    xs: 4px
    sm: 8px
    md: 16px
    lg: 24px
    xl: 32px

colors:
  light:
    background:
      primary: '#ffffff'
      secondary: '#f9fafb'
      tertiary: '#f3f4f6'
    text:
      primary: '#111827'
      secondary: '#6b7280'
      muted: '#9ca3af'
    border:
      default: '#e5e7eb'
      emphasis: '#d1d5db'
      focus: '#3b82f6'
    status:
      positive: '#16a34a'
      negative: '#dc2626'
      warning: '#d97706'
      info: '#0891b2'
    interactive:
      primary: '#3b82f6'
      primaryHover: '#2563eb'
      primaryActive: '#1d4ed8'
      primaryText: '#ffffff'
      disabled: '#f3f4f6'
      disabledText: '#9ca3af'
    category:
      executive: '#8b5cf6'
      rules: '#0891b2'
      notices: '#f59e0b'
      presidential: '#dc2626'

typography:
  fontFamily: 'Inter, -apple-system, BlinkMacSystemFont, "Segoe UI", system-ui, sans-serif'
  fontMono: '"JetBrains Mono", "Fira Code", "Courier New", monospace'
  sizes:
    xs: 12px
    sm: 14px
    base: 16px
    lg: 18px
    xl: 20px
    2xl: 24px
    3xl: 30px
    4xl: 36px
  weights:
    normal: 400
    medium: 500
    semibold: 600
    bold: 700

spacing:
  xs: 4px
  sm: 8px
  md: 16px
  lg: 24px
  xl: 32px
  2xl: 48px

borders:
  width:
    thin: 1px
    medium: 2px
  radius: 4px
  style: solid

shadows:
  subtle: 0 1px 2px rgba(0, 0, 0, 0.05)
  sm: 0 1px 3px rgba(0, 0, 0, 0.1)
  md: 0 4px 6px rgba(0, 0, 0, 0.1)
  lg: 0 10px 15px rgba(0, 0, 0, 0.1)

animations:
  duration:
    fast: 100ms
    base: 150ms
    slow: 300ms
  easing: cubic-bezier(0.4, 0, 0.2, 1)

breakpoints:
  xs: 0-639px
  sm: 640-767px
  md: 768-1023px
  lg: 1024-1279px
  xl: 1280px+

components:
  buttons:
    padding: 10px 16px
    minHeight: 40px
    border: 1px solid
    borderRadius: 4px
    fontSize: sm
    fontWeight: medium
    transition: all 150ms ease-in-out
    variants:
      primary:
        bg: '{colors.light.interactive.primary}'
        color: '#ffffff'
      secondary:
        bg: '{colors.light.background.secondary}'
        color: '{colors.light.text.primary}'
        border: '{colors.light.border.default}'
      ghost:
        bg: transparent
        color: '{colors.light.interactive.primary}'
      danger:
        bg: '{colors.light.status.negative}'
        color: '#ffffff'

  inputs:
    padding: 10px 12px
    height: 40px
    border: 1px solid
    borderColor: '{colors.light.border.default}'
    borderRadius: 4px
    fontSize: sm
    focus:
      borderColor: '{colors.light.border.focus}'
      boxShadow: '0 0 0 3px rgba(59, 130, 246, 0.1)'

  cards:
    padding: 16px
    border: 1px solid
    borderColor: '{colors.light.border.default}'
    borderRadius: 6px
    shadow: sm

  badges:
    padding: 4px 8px
    fontSize: xs
    fontWeight: medium
    borderRadius: 4px

  tables:
    cellPadding: 12px
    headerBg: '{colors.light.background.tertiary}'
    rowBorder: 1px solid
    rowHover:
      bg: '{colors.light.background.secondary}'

shadcnComponents:
  availability: All components imported from @/components/ui
  philosophy: Use shadcn components as the primary UI primitive. Only build custom components when shadcn doesn't provide the pattern.
  components:
    Button:
      use: All interactive buttons and CTAs
      variants: [default, destructive, outline, secondary, ghost, link]
      sizes: [default, sm, lg, icon]
      accessibility: Includes focus-visible ring, disabled state handling
      example: "CTA buttons, form submissions, navigation actions"
    
    Card:
      use: Content containers, article cards, data displays
      structure: Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter
      accessibility: Semantic HTML with proper heading hierarchy
      example: "Article preview cards, document containers, information panels"
    
    Badge:
      use: Document category labels, status indicators, tags
      guidance: "Use semantic colors (category.executive, category.rules, etc.) for federal document types"
      example: "Tagging articles as 'Rules', 'Notices', 'Executive Orders'"
    
    Input:
      use: Search bars, form fields, text inputs
      accessibility: Proper focus states and disabled handling
      example: "Search Federal Register input field"
    
    Alert:
      use: Error messages, success confirmations, informational messages
      variants: [default, destructive] - use with status colors
      accessibility: Role=alert for important messages
      example: "Failed to load articles, success messages"
    
    Skeleton:
      use: Loading placeholders while fetching data
      guidance: "Use for article cards, feed lists during data fetch"
      example: "Card skeleton while articles load"

patterns:
  navigation:
    style: Horizontal with text labels
    active: Primary color or underline
  
  documentFeed:
    layout: Vertical card stack
    spacing: md
    cardComponent: Use Card + CardContent with responsive grid
  
  searchBar:
    placement: Sticky header
    width: Full or constrained
    component: Input with icon wrapper
  
  alerts:
    success: Alert variant with status.positive color
    warning: Alert variant with status.warning color
    error: Alert variant with status.negative color (destructive variant)
    dismissible: Include close button in Alert component
  
  articleCard:
    component: Card with CardHeader, CardContent, CardFooter
    footer: Link and external link buttons using Button component
    status: Use Badge for document category
  
  formSubmit:
    component: Button variant=default with proper loading state
    disabled: Managed by Button component automatically

accessibility:
  wcagLevel: AA
  contrast:
    minRatio: 4.5:1
    largeText: 3:1
  focusIndicators:
    style: 2px solid
    offset: 2px
    always: visible
  keyboard:
    allInteractive: accessible
    shortcuts:
      Escape: Close modals
      Enter: Submit forms
      Tab: Navigate forward
      Shift+Tab: Navigate backward
  testing:
    tools:
      - axe-core
      - WebAIM contrast checker
      - Keyboard navigation testing
      - Screen reader testing (NVDA, VoiceOver)

responsive:
  approach: Mobile-first
  xs:
    columns: 1
    padding: 16px
  md:
    columns: 2
    padding: 20px
  lg:
    columns: 3
    padding: 24px
  xl:
    columns: 3+
    padding: 32px

guidelines:
  do:
    - Use semantic color for status and category
    - Pair icons with text labels
    - Ensure 44x44px minimum touch targets
    - Test keyboard and screen reader access
    - Use plain language
    - Show feedback on interactions
    - Optimize images (WebP, lazy load)
    - Category-code documents with color badges
  
  dont:
    - Use color alone for meaning
    - Auto-play audio/video
    - Require mouse-only interaction
    - Use text smaller than 14px (body)
    - Hide interactive elements
    - Use jargon without explanation
    - Nest modals
    - Load unoptimized images

migration:
  description: "All components now use shadcn primitives. This guide explains what's changed."
  replaced:
    customButtons: "Use Button component with variants: default, destructive, outline, secondary, ghost, link"
    customCards: "Use Card component with CardHeader, CardTitle, CardDescription, CardContent, CardFooter"
    customAlerts: "Use Alert component with variant destructive for errors"
    customInputs: "Use Input component - handles all focus states and styling"
    customSkeletons: "Use Skeleton component for loading placeholders"
  comingNext:
    - "Replace remaining custom divs with shadcn components"
    - "Leverage Button variants for all interactive elements"
    - "Use Badge for document category labels and status indicators"

reference:
   inspiration:
     - Apple Design System
     - Stripe
     - Gov.uk
   targetAudience: Americans seeking government transparency
   philosophy: Make government data accessible and actionable for everyday citizens
