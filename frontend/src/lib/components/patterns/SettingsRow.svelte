<script lang="ts">
	import type { Snippet } from 'svelte';
	import './patterns.css';

	export type SettingsRowControl = { labelId: string; descriptionId: string | undefined };

	interface Props {
		label: string;
		description?: string;
		control: Snippet<[SettingsRowControl]>;
		disabled?: boolean;
		class?: string;
	}

	let { label, description, control, disabled = false, class: className }: Props = $props();

	const baseId = $props.id();
	const labelId = `${baseId}-label`;
	const descriptionId = $derived(description ? `${baseId}-description` : undefined);
</script>

<div class="atelier-settings-row {className ?? ''}" data-disabled={disabled || undefined}>
	<div class="atelier-settings-row__copy">
		<span id={labelId} class="atelier-settings-row__label">{label}</span>
		{#if description}<span id={descriptionId} class="atelier-settings-row__description">{description}</span>{/if}
	</div>
	<div class="atelier-settings-row__control">{@render control({ labelId, descriptionId })}</div>
</div>
