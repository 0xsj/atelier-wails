<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { HTMLAttributes } from 'svelte/elements';
	import './forms.css';

	type FieldMessageTone = 'hint' | 'error' | 'success';

	interface Props extends Omit<HTMLAttributes<HTMLParagraphElement>, 'children' | 'class' | 'role'> { tone?: FieldMessageTone; children?: Snippet; class?: string; }

	let { tone = 'hint', children, class: className, ...rest }: Props = $props();
</script>

<p {...rest} class="atelier-field-message {className ?? ''}" data-tone={tone} role={tone === 'error' ? 'alert' : 'status'}>{@render children?.()}</p>
