<script lang="ts">
	import type { Snippet } from 'svelte';
	import ChoiceDialog, { type ChoiceAction } from './ChoiceDialog.svelte';

	export type ConfirmOutcome = 'confirm' | 'cancel';
	type ConfirmTone = 'default' | 'danger';

	interface Props {
		open?: boolean;
		title: string;
		description?: string;
		confirmLabel?: string;
		cancelLabel?: string;
		tone?: ConfirmTone;
		busy?: boolean;
		onResolve?: (outcome: ConfirmOutcome) => void;
		children?: Snippet;
		class?: string;
	}

	let { open = $bindable(false), title, description, confirmLabel = 'Confirm', cancelLabel = 'Cancel', tone = 'default', busy = false, onResolve, children, class: className }: Props = $props();

	const actions = $derived<readonly ChoiceAction[]>([
		{ id: 'cancel', label: cancelLabel, variant: 'secondary' },
		{ id: 'confirm', label: confirmLabel, variant: tone === 'danger' ? 'danger' : 'primary', primary: true }
	]);
</script>

<ChoiceDialog bind:open {title} {description} {actions} {tone} {busy} onResolve={(outcome) => onResolve?.(outcome === 'confirm' ? 'confirm' : 'cancel')} {children} class={className} />
