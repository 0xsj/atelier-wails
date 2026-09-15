<script lang="ts">
	import type { HTMLInputAttributes } from 'svelte/elements';
	import './forms.css';

	interface Props extends Omit<HTMLInputAttributes, 'class' | 'value' | 'type' | 'size'> {
		value?: string;
		placeholder?: string;
		size?: 'sm' | 'md' | 'lg';
		ariaLabel?: string;
		class?: string;
	}

	let { value = $bindable(''), placeholder = 'Search', size = 'md', ariaLabel = 'Search', class: className, ...rest }: Props = $props();
</script>

<div class="atelier-search-field {className ?? ''}" data-size={size}><span class="atelier-search-field__icon" aria-hidden="true">⌕</span><input {...rest} bind:value type="search" class="atelier-search-field__input" placeholder={placeholder} aria-label={ariaLabel} /></div>
