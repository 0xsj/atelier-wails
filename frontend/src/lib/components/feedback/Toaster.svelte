<script lang="ts">
	import type { HTMLAttributes } from 'svelte/elements';
	import Toast, { type ToastItem } from './Toast.svelte';
	import './feedback.css';

	interface Props extends Omit<HTMLAttributes<HTMLElement>, 'class' | 'role'> {
		toasts: readonly ToastItem[];
		onDismiss?: (id: string) => void;
		class?: string;
	}

	let { toasts, onDismiss, class: className, ...rest }: Props = $props();
</script>

<aside {...rest} class="atelier-toaster {className ?? ''}" aria-label="Notifications">
	{#each toasts as toast (toast.id)}<Toast {...toast} onDismiss={() => onDismiss?.(toast.id)} />{/each}
</aside>
