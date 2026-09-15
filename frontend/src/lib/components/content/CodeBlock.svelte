<script lang="ts">
	import type { HTMLAttributes } from 'svelte/elements';
	import CopyButton from './CopyButton.svelte';
	import './content.css';

	interface Props extends Omit<HTMLAttributes<HTMLElement>, 'children' | 'class'> {
		code: string;
		language?: string;
		copyable?: boolean;
		onCopy?: (value: string) => void;
		class?: string;
	}

	let { code, language, copyable = true, onCopy, class: className, ...rest }: Props = $props();
</script>

<figure {...rest} class="atelier-code-block {className ?? ''}">
	<header class="atelier-code-block__header">
		{#if language}<span>{language}</span>{:else}<span>Code</span>{/if}
		{#if copyable}<CopyButton value={code} onCopy={onCopy} />{/if}
	</header>
	<pre><code>{code}</code></pre>
</figure>
