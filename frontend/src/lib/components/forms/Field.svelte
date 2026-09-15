<script lang="ts">
	import type { Snippet } from 'svelte';
	import { Label } from '../primitives';
	import FieldMessage from './FieldMessage.svelte';
	import './forms.css';

	type FieldTone = 'hint' | 'error' | 'success';

	export type FieldControl = {
		id: string;
		describedBy: string | undefined;
		invalid: boolean;
	};

	interface Props {
		label: string;
		description?: string;
		message?: string;
		tone?: FieldTone;
		required?: boolean;
		disabled?: boolean;
		children: Snippet<[FieldControl]>;
		class?: string;
	}

	let { label, description, message, tone = 'hint', required = false, disabled = false, children, class: className }: Props = $props();

	const baseId = $props.id();
	const controlId = `${baseId}-control`;
	const descriptionId = `${baseId}-description`;
	const messageId = `${baseId}-message`;
	const describedBy = $derived([description ? descriptionId : null, message ? messageId : null].filter(Boolean).join(' ') || undefined);
	const invalid = $derived(tone === 'error' && Boolean(message));
</script>

<div class="atelier-field {className ?? ''}" data-disabled={disabled || undefined} data-invalid={invalid || undefined}>
	<div class="atelier-field__heading">
		<Label for={controlId}>{label}{#if required}<span class="atelier-field__required" aria-hidden="true">*</span>{/if}</Label>
		{#if description}<p id={descriptionId} class="atelier-field__description">{description}</p>{/if}
	</div>
	<div class="atelier-field__control">{@render children({ id: controlId, describedBy, invalid })}</div>
	{#if message}<FieldMessage id={messageId} {tone}>{message}</FieldMessage>{/if}
</div>
