<script lang="ts">
	import type { HTMLAttributes } from 'svelte/elements';
import './feedback.css';

	type ToastTone = 'neutral' | 'accent' | 'success' | 'warning' | 'danger';

	export type ToastItem = { id: string; title: string; description?: string; tone?: ToastTone; actionLabel?: string };

	interface Props extends Omit<HTMLAttributes<HTMLDivElement>, 'class' | 'role'> {
		title: string;
		description?: string;
		tone?: ToastTone;
		actionLabel?: string;
		onAction?: () => void;
		onDismiss?: () => void;
		class?: string;
	}

	let { title, description, tone = 'neutral', actionLabel, onAction, onDismiss, class: className, ...rest }: Props = $props();
</script>

<div {...rest} class="atelier-toast {className ?? ''}" data-tone={tone} role={tone === 'danger' ? 'alert' : 'status'}>
	<span class="atelier-toast__marker" aria-hidden="true"></span>
	<div class="atelier-toast__copy"><strong>{title}</strong>{#if description}<span>{description}</span>{/if}</div>
	{#if actionLabel}<button type="button" class="atelier-toast__action" onclick={onAction}>{actionLabel}</button>{/if}
	<button type="button" class="atelier-toast__close" aria-label="Dismiss notification" onclick={onDismiss}>×</button>
</div>
