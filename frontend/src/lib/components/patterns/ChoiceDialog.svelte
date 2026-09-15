<script lang="ts">
	import type { Snippet } from 'svelte';
	import { AlertDialog as Primitive } from 'bits-ui';
	import { Button, type ButtonVariant } from '../primitives';
	import './patterns.css';

	export type ChoiceAction = {
		id: string;
		label: string;
		variant?: ButtonVariant;
		side?: 'start' | 'end';
		primary?: boolean;
		key?: string;
		disabled?: boolean;
	};
	type ChoiceTone = 'default' | 'danger';

	interface Props {
		open?: boolean;
		title: string;
		description?: string;
		actions: readonly ChoiceAction[];
		cancelId?: string;
		tone?: ChoiceTone;
		busy?: boolean;
		onResolve?: (outcome: string) => void;
		children?: Snippet;
		class?: string;
	}

	let { open = $bindable(false), title, description, actions, cancelId = 'cancel', tone = 'default', busy = false, onResolve, children, class: className }: Props = $props();

	let settled = false;
	const startActions = $derived(actions.filter((action) => action.side === 'start'));
	const endActions = $derived(actions.filter((action) => action.side !== 'start'));

	$effect(() => {
		if (open) settled = false;
	});

	function settle(outcome: string): void {
		if (settled) return;
		settled = true;
		onResolve?.(outcome);
	}

	function choose(action: ChoiceAction): void {
		if (action.disabled || busy) return;
		settle(action.id);
		if (action.id === cancelId) open = false;
	}

	function handleOpenChange(next: boolean): void {
		if (!next) settle(cancelId);
	}

	function matchesKey(event: KeyboardEvent, key: string | undefined): boolean {
		if (!key) return false;
		const parts = key.toLowerCase().split('+');
		const wantsMod = parts.includes('mod');
		const wantsShift = parts.includes('shift');
		const letter = parts[parts.length - 1];
		return (wantsMod ? event.metaKey || event.ctrlKey : !event.metaKey && !event.ctrlKey) && event.shiftKey === wantsShift && event.key.toLowerCase() === letter;
	}

	function handleKeydown(event: KeyboardEvent): void {
		const target = actions.find((action) => matchesKey(event, action.key));
		if (!target) return;
		event.preventDefault();
		choose(target);
	}
</script>

<Primitive.Root bind:open onOpenChange={handleOpenChange}>
	<Primitive.Portal>
		<Primitive.Overlay class="atelier-confirm__overlay" />
		<Primitive.Content class="atelier-confirm {className ?? ''}" data-tone={tone} escapeKeydownBehavior={busy ? 'ignore' : 'close'} interactOutsideBehavior={busy ? 'ignore' : 'close'} onkeydown={handleKeydown}>
			<div class="atelier-confirm__heading">
				<Primitive.Title class="atelier-confirm__title">{title}</Primitive.Title>
				{#if description}<Primitive.Description class="atelier-confirm__description">{description}</Primitive.Description>{/if}
			</div>
			{#if children}<div class="atelier-confirm__body">{@render children()}</div>{/if}
			<div class="atelier-confirm__actions" data-split={startActions.length ? true : undefined}>
				{#if startActions.length}<div class="atelier-confirm__actions-group">{#each startActions as action (action.id)}<Button variant={action.variant ?? 'secondary'} disabled={busy || action.disabled} onclick={() => choose(action)}>{action.label}</Button>{/each}</div>{/if}
				<div class="atelier-confirm__actions-group">
					{#each endActions as action (action.id)}
						<!-- svelte-ignore a11y_autofocus -->
						<Button variant={action.variant ?? (action.primary ? 'primary' : 'secondary')} disabled={(busy && !action.primary) || action.disabled} loading={busy && action.primary} autofocus={action.primary || undefined} onclick={() => choose(action)}>{action.label}</Button>
					{/each}
				</div>
			</div>
		</Primitive.Content>
	</Primitive.Portal>
</Primitive.Root>
