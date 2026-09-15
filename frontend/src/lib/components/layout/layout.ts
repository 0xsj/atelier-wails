export const LAYOUT_GAPS = ['none', 'xs', 'sm', 'md', 'lg', 'xl'] as const;
export type LayoutGap = (typeof LAYOUT_GAPS)[number];

export type LayoutAlign = 'stretch' | 'start' | 'center' | 'end' | 'baseline';
export type LayoutJustify = 'start' | 'center' | 'end' | 'between';
export type LayoutTag = 'div' | 'section' | 'article' | 'nav' | 'ul' | 'ol' | 'li' | 'header' | 'footer';
