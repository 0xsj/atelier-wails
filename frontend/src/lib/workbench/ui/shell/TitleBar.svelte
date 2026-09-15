<script lang="ts">
	import type { Snippet } from 'svelte';
	import './shell.css';

	type TitleBarInset = 'none' | 'macos' | 'windows';

	interface Props {
		title: string;
		subtitle?: string;
		inset?: TitleBarInset;
		active?: boolean;
		dragRegion?: Record<string, string>;
		dragExclude?: Record<string, string>;
		leading?: Snippet;
		center?: Snippet;
		trailing?: Snippet;
		class?: string;
	}

	let { title, subtitle, inset = 'none', active = true, dragRegion = {}, dragExclude = {}, leading, center, trailing, class: className }: Props = $props();
</script>

<header class="atelier-title-bar {className ?? ''}" data-inset={inset} data-active={active} {...dragRegion}>
	<div class="atelier-title-bar__side" {...dragRegion}>{#if leading}<div class="atelier-title-bar__controls" {...dragExclude}>{@render leading()}</div>{/if}</div>
	<div class="atelier-title-bar__center" {...dragRegion}>
		{#if center}
			<div class="atelier-title-bar__controls" {...dragExclude}>{@render center()}</div>
		{:else}
			<span class="atelier-title-bar__title" {...dragRegion}>{title}</span>
			{#if subtitle}<span class="atelier-title-bar__subtitle" {...dragRegion}>{subtitle}</span>{/if}
		{/if}
	</div>
	<div class="atelier-title-bar__side" data-side="end" {...dragRegion}>{#if trailing}<div class="atelier-title-bar__controls" {...dragExclude}>{@render trailing()}</div>{/if}</div>
</header>
