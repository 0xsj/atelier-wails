<script lang="ts">
	import type { HTMLButtonAttributes } from 'svelte/elements';
	import './content.css';

	interface Props extends Omit<HTMLButtonAttributes, 'children' | 'class' | 'onclick'> {
		value: string;
		label?: string;
		onCopy?: (value: string) => void;
		class?: string;
	}

	let { value, label = 'Copy', onCopy, class: className, ...rest }: Props = $props();
	let copied = $state(false);
	let resetTimer: ReturnType<typeof setTimeout> | undefined;

	async function copy(): Promise<void> {
		if (!navigator.clipboard) return;
		try {
			await navigator.clipboard.writeText(value);
			copied = true;
			onCopy?.(value);
			if (resetTimer) window.clearTimeout(resetTimer);
			resetTimer = window.setTimeout(() => (copied = false), 1600);
		} catch {
			copied = false;
		}
	}
</script>

<button {...rest} type="button" class="atelier-copy-button {className ?? ''}" aria-label={copied ? 'Copied' : label} onclick={copy}>
	<span aria-hidden="true">{copied ? '✓' : '⧉'}</span>{copied ? 'Copied' : label}
</button>
