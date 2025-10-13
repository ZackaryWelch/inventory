# Visual Verification Checklist

## Instructions

For each screen, compare React (left) vs Go WASM (right) side-by-side.

**Rating Scale:**
- ✅ **Perfect**: Pixel-perfect match (< 1% difference)
- ⚠️ **Close**: Visually similar with minor differences (1-5% difference)
- ❌ **Fail**: Significant visual differences (> 5% difference)
- ⏭️ **Skip**: Screen not yet implemented

## 1. Authentication Screens

### 1.1 Login Screen (ViewLogin)

**Layout**
- [ ] Centered container (flex items-center justify-center h-screen)
- [ ] Vertical column layout
- [ ] Logo placeholder (w-32 h-26 mb-20)

**Typography**
- [ ] App title: "Nishiki Inventory" (text-2xl, 24px, bold)
- [ ] Subtitle: "Inventory Management System" (text-sm, 14px, gray)

**Button**
- [ ] Primary button styling (bg-primary, text-white)
- [ ] Text: "Sign In with Authentik"
- [ ] Large size (h-12, 48px height)
- [ ] Rounded corners (rounded, 10px)
- [ ] Proper padding (px-6 py-3)

**Spacing**
- [ ] Vertical gap between elements (gap-6, 24px)
- [ ] Logo margin bottom (mb-20, 80px)

**Colors**
- [ ] Background: #f9fafb (gray-lightest)
- [ ] Button: #6ab3ab (primary)
- [ ] Text: #000000 (black)
- [ ] Subtitle: #4b5563 (gray-dark)

**Overall Rating:** [ ]

---

### 1.2 Callback Screen (ViewCallback)

**Layout**
- [ ] Centered container (min-h-screen flex items-center justify-center)
- [ ] Text center alignment

**Spinner**
- [ ] Loading spinner present
- [ ] Size: h-12 w-12 (48px)
- [ ] Border color and width
- [ ] Center alignment with mx-auto
- [ ] Margin bottom: mb-4 (16px)

**Typography**
- [ ] Title: "Completing Sign In..." (text-2xl, 24px)
- [ ] Message: authentication message (text-sm, 14px, gray)

**Colors**
- [ ] Background: #f9fafb (gray-lightest)
- [ ] Spinner border: #6ab3ab (primary)

**Overall Rating:** [ ]

---

## 2. Dashboard Screen (ViewDashboard)

### 2.1 Header

**Layout**
- [ ] Full width (w-full)
- [ ] Height: h-12 (48px)
- [ ] Flex row with space-between
- [ ] Items center aligned
- [ ] Horizontal padding: px-4 (16px)

**Typography**
- [ ] Title: "Dashboard" (text-xl, 20px, semibold)
- [ ] Username button text (text-base, 16px)

**Colors**
- [ ] Background: #ffffff (white)
- [ ] Title text: #000000 (black)

**Overall Rating:** [ ]

---

### 2.2 Navigation Buttons

**Layout**
- [ ] Flex row with wrap
- [ ] Gap: gap-3 (12px)
- [ ] Max width: max-w-lg (512px)
- [ ] Centered with mx-auto

**Button Grid**
- [ ] 4 buttons visible: Groups, Collections, Profile, Search
- [ ] Equal sizing
- [ ] Proper icon + text layout

**Individual Buttons**
- [ ] Primary variant styling
- [ ] Icon size: 24px
- [ ] Icon color: #ffffff (white)
- [ ] Background: #6ab3ab (primary)
- [ ] Height: h-10 (40px) for medium size
- [ ] Border radius: rounded (10px)
- [ ] Padding: px-4 py-2 (16px horizontal, 8px vertical)

**Icons**
- [ ] Groups: Group icon
- [ ] Collections: FolderOpen icon
- [ ] Profile: Person icon
- [ ] Search: Search icon

**Overall Rating:** [ ]

---

### 2.3 Stats Section

**Layout**
- [ ] Column direction (flex-col)
- [ ] Background: #ffffff (white)
- [ ] Border radius: rounded (10px)
- [ ] Padding: p-4 (16px)
- [ ] Gap: gap-3 (12px)

**Stats Title**
- [ ] Text: "Quick Stats"
- [ ] Font size: text-lg (18px)
- [ ] Font weight: semibold

**Stats Grid**
- [ ] Flex row with wrap
- [ ] Gap: gap-4 (16px)

**Stat Cards**
- [ ] Two cards: Groups count, Collections count
- [ ] Groups card background: #6ab3ab (primary) tint
- [ ] Collections card background: #fcd884 (accent) tint
- [ ] Border radius: rounded (10px)
- [ ] Padding: p-3 (12px)

**Stat Card Content**
- [ ] Value text: large, bold
- [ ] Label text: text-sm (14px), gray

**Overall Rating:** [ ]

---

## 3. Groups Screen (ViewGroups)

### 3.1 Header

**Layout**
- [ ] Full width header
- [ ] Back button present (ArrowBack icon)
- [ ] Title: "Groups"
- [ ] Left-aligned content

**Back Button**
- [ ] Rounded background: bg-gray-light
- [ ] Padding: p-2 (8px)
- [ ] Icon size: 20px
- [ ] Border radius: rounded-full (9999px)

**Overall Rating:** [ ]

---

### 3.2 Create Button

**Layout**
- [ ] Primary button variant
- [ ] Text: "Create Group"
- [ ] Icon: Add icon (plus)
- [ ] Medium size (h-10, 40px)

**Positioning**
- [ ] Top of content area
- [ ] Proper margin bottom

**Overall Rating:** [ ]

---

### 3.3 Groups List

**Empty State**
- [ ] Message: "No groups found. Create your first group!"
- [ ] Text center aligned
- [ ] Padding: p-8 (32px)
- [ ] Text color: #4b5563 (gray-dark)

**Group Cards (if data present)**
- [ ] Card layout: flex justify-between gap-2
- [ ] Background: #ffffff (white)
- [ ] Border radius: rounded (10px)
- [ ] Padding applied correctly

**Group Card Content**
- [ ] Group name: text-lg (18px), leading-6 (24px)
- [ ] Member count: text-sm (14px), gray
- [ ] Menu button: MoreVert icon, ghost variant

**Overall Rating:** [ ]

---

## 4. Collections Screen (ViewCollections)

### 4.1 Header

**Layout**
- [ ] Same as Groups header
- [ ] Title: "Collections"
- [ ] Back button present

**Overall Rating:** [ ]

---

### 4.2 Create Button

**Layout**
- [ ] Primary button variant
- [ ] Text: "Create Collection"
- [ ] Icon: Add icon
- [ ] Medium size

**Overall Rating:** [ ]

---

### 4.3 Collections List

**Empty State**
- [ ] Message: "No collections found. Create your first collection!"
- [ ] Styling same as Groups empty state

**Collection Cards (if data present)**
- [ ] Card layout: flex justify-between gap-2
- [ ] Icon circle: bg-accent, rounded-full, w-11 h-11 (44px)
- [ ] Icon inside circle: size 24px, black color
- [ ] Collection name: leading-5 (20px)
- [ ] Menu button: MoreVert icon

**Overall Rating:** [ ]

---

## 5. Profile Screen (ViewProfile)

### 5.1 Header

**Layout**
- [ ] Same as Groups/Collections header
- [ ] Title: "Profile"
- [ ] Back button present

**Overall Rating:** [ ]

---

### 5.2 User Info Card

**Layout**
- [ ] Card component with white background
- [ ] Border radius: rounded (10px)
- [ ] Padding: p-4 (16px)

**Fields**
- [ ] Username label and value
- [ ] Email label and value
- [ ] Name label and value (if present)
- [ ] Labels: bold or semibold
- [ ] Values: normal weight

**Spacing**
- [ ] Proper gap between field groups

**Overall Rating:** [ ]

---

### 5.3 Logout Button

**Layout**
- [ ] Danger button variant
- [ ] Text: "Sign Out"
- [ ] Icon: Logout icon
- [ ] Medium size
- [ ] Red background: #cd5a5a (danger)
- [ ] White text

**Positioning**
- [ ] Below user info card
- [ ] Proper margin top

**Overall Rating:** [ ]

---

### 5.4 Developer Tools Section

**Layout**
- [ ] Column direction
- [ ] Background: #ffffff (white)
- [ ] Border radius: rounded (10px)
- [ ] Padding: p-4 (16px)
- [ ] Gap: gap-3 (12px)
- [ ] Margin top: mt-4 (16px)

**Title**
- [ ] Text: "Developer Tools"
- [ ] Font size and weight appropriate

**Clear Cache Button**
- [ ] Primary button variant
- [ ] Text: "Clear Cache & Reload"
- [ ] Icon: Refresh icon
- [ ] Medium size

**Overall Rating:** [ ]

---

## 6. Cross-Screen Consistency

### 6.1 Typography System

- [ ] H1/App Title: 24px (text-2xl), bold
- [ ] H2/Section Title: 20px (text-xl), semibold
- [ ] H3/Card Title: 18px (text-lg), semibold
- [ ] Body text: 16px (text-base)
- [ ] Small text: 14px (text-sm)
- [ ] Micro text: 12px (text-xs)

---

### 6.2 Color Palette

**Primary Colors**
- [ ] Primary: #6ab3ab
- [ ] Primary Light: #95cec6
- [ ] Primary Lightest: #d6eae7
- [ ] Primary Dark: #558f89

**Accent Colors**
- [ ] Accent: #fcd884
- [ ] Accent Dark: #f2c04e

**Danger Colors**
- [ ] Danger: #cd5a5a
- [ ] Danger Dark: #b84848

**Gray Scale**
- [ ] Gray Lightest: #f9fafb
- [ ] Gray Light: #e5e7eb
- [ ] Gray: #9ca3af
- [ ] Gray Dark: #4b5563

**Base Colors**
- [ ] White: #ffffff
- [ ] Black: #000000

---

### 6.3 Spacing System

**Consistency Checks**
- [ ] gap-1 (4px) used correctly
- [ ] gap-2 (8px) used correctly
- [ ] gap-3 (12px) used correctly
- [ ] gap-4 (16px) used correctly
- [ ] gap-6 (24px) used correctly

**Padding Checks**
- [ ] p-2 (8px) used correctly
- [ ] p-3 (12px) used correctly
- [ ] p-4 (16px) used correctly
- [ ] p-6 (24px) used correctly
- [ ] p-8 (32px) used correctly

---

### 6.4 Border Radius

- [ ] rounded (10px) for cards and containers
- [ ] rounded-full (9999px) for icon circles and pills
- [ ] rounded-lg (8px) for medium components
- [ ] rounded-2xl (16px) for large containers

---

### 6.5 Component Library

**Buttons**
- [ ] Primary variant consistent
- [ ] Danger variant consistent
- [ ] Cancel variant consistent
- [ ] Ghost variant consistent
- [ ] Icon buttons consistent

**Cards**
- [ ] Base card styling consistent
- [ ] Flex-between pattern correct
- [ ] Header/Content separation clear

**Inputs**
- [ ] Base input styling consistent
- [ ] Rounded variant consistent
- [ ] Search variant consistent
- [ ] Form field layout consistent

**Icons**
- [ ] Icon sizes consistent (16px, 20px, 24px)
- [ ] Icon colors match design tokens
- [ ] Icon circles render correctly

---

## 7. Responsive Behavior

### 7.1 Mobile (375px)
- [ ] All content visible without horizontal scroll
- [ ] Buttons tap-friendly (min 44px height)
- [ ] Text readable without zoom
- [ ] Spacing appropriate for small screens

### 7.2 Mobile Large (414px)
- [ ] Content scales appropriately
- [ ] No awkward empty space
- [ ] Buttons and cards maintain proportions

---

## 8. Interaction States

### 8.1 Hover States
- [ ] Buttons show hover feedback
- [ ] Cards show hover feedback (if clickable)
- [ ] Icons show hover feedback

### 8.2 Active States
- [ ] Buttons show active/pressed state
- [ ] Navigation items show active state

### 8.3 Focus States
- [ ] Keyboard navigation works
- [ ] Focus indicators visible
- [ ] Tab order logical

### 8.4 Disabled States
- [ ] Disabled buttons show correct styling
- [ ] Disabled inputs show correct styling
- [ ] Cursor indicates disabled state

---

## 9. Loading States

### 9.1 Loading Spinner
- [ ] Spinner renders correctly
- [ ] Animation smooth (if supported)
- [ ] Size appropriate (48px)

### 9.2 Loading Skeleton (if implemented)
- [ ] Skeleton placeholders render
- [ ] Animation present
- [ ] Layout matches final content

---

## 10. Error States

### 10.1 Empty States
- [ ] Empty state component renders
- [ ] Message clear and helpful
- [ ] Styling matches design

### 10.2 Error Messages (if implemented)
- [ ] Error text visible
- [ ] Error color: danger red
- [ ] Icon present (if applicable)

---

## Summary

### Overall Statistics

**Total Screens Verified:** ___ / 10
**Perfect Match (✅):** ___
**Close Match (⚠️):** ___
**Failed (❌):** ___
**Not Implemented (⏭️):** ___

**Overall Visual Parity:** ____%

### Critical Issues Found

1. ________________________________________________
2. ________________________________________________
3. ________________________________________________

### Minor Issues Found

1. ________________________________________________
2. ________________________________________________
3. ________________________________________________

### Recommendations

1. ________________________________________________
2. ________________________________________________
3. ________________________________________________

### Sign-off

**Verified By:** ________________
**Date:** ________________
**Status:** [ ] PASS [ ] FAIL [ ] NEEDS FIXES

**Next Steps:**
- [ ] Document all issues in VISUAL_ISSUES.md
- [ ] Create tickets for critical issues
- [ ] Schedule fix iteration
- [ ] Plan re-verification date
