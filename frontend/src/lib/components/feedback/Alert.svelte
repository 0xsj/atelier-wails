<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { HTMLAttributes } from 'svelte/elements';
	import './feedback.css';

	type Tone = 'neutral' | 'accent' | 'success' | 'warning' | 'danger';

	interface Props extends Omit<HTMLAttributes<HTMLDivElement>, 'children' | 'class' | 'role'> {
		tone?: Tone;
		title?: string;
		role?: 'status' | 'alert';
		dismissible?: boolean;
		children?: Snippet;
		onDismiss?: () => void;
		class?: string;
	}

	let { tone = 'neutral', title, role, dismissible = false, children, onDismiss, class: className, ...rest }: Props = $props();
	const liveRole = $derived(role ?? (tone === 'danger' || tone === 'warning' ? 'alert' : 'status'));
</script>

<div {...rest} class="atelier-alert {className ?? ''}" data-tone={tone} role={liveRole}>
	<span class="atelier-alert__marker" aria-hidden="true"></span>
	<div class="atelier-alert__content">{#if title}<strong>{title}</strong>{/if}{#if children}<div class="atelier-alert__message">{@render children()}</div>{/if}</div>
	{#if dismissible}<button type="button" class="atelier-alert__close" aria-label="Dismiss alert" onclick={onDismiss}>×</button>{/if}
</div>
