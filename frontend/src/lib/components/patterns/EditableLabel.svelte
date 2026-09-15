<script lang="ts">
	import { tick } from 'svelte';
	import './patterns.css';

	type EditableSize = 'sm' | 'md' | 'lg';

	interface Props {
		value?: string;
		placeholder?: string;
		ariaLabel: string;
		size?: EditableSize;
		disabled?: boolean;
		allowEmpty?: boolean;
		onCommit?: (value: string) => void;
		class?: string;
	}

	let { value = $bindable(''), placeholder = 'Untitled', ariaLabel, size = 'md', disabled = false, allowEmpty = false, onCommit, class: className }: Props = $props();

	let editing = $state(false);
	let draft = $state('');
	let trigger = $state<HTMLButtonElement | null>(null);
	let input = $state<HTMLInputElement | null>(null);

	async function beginEdit(): Promise<void> {
		if (disabled) return;
		draft = value;
		editing = true;
		await tick();
		input?.select();
	}

	async function finish(): Promise<void> {
		editing = false;
		await tick();
		trigger?.focus();
	}

	function commit(): void {
		if (!editing) return;
		const next = draft.trim();
		if (next !== value && (next || allowEmpty)) {
			value = next;
			onCommit?.(next);
		}
		void finish();
	}

	function cancel(): void {
		if (!editing) return;
		void finish();
	}

	function handleTriggerKeydown(event: KeyboardEvent): void {
		if (event.key === 'Enter' || event.key === 'F2') {
			event.preventDefault();
			void beginEdit();
		}
	}

	function handleInputKeydown(event: KeyboardEvent): void {
		if (event.key === 'Enter') {
			event.preventDefault();
			commit();
		} else if (event.key === 'Escape') {
			event.preventDefault();
			cancel();
		}
	}
</script>

<span class="atelier-editable-label {className ?? ''}" data-size={size} data-editing={editing || undefined}>
	{#if editing}
		<!-- svelte-ignore a11y_autofocus -->
		<input bind:this={input} bind:value={draft} class="atelier-editable-label__input" type="text" aria-label={ariaLabel} {placeholder} autofocus onkeydown={handleInputKeydown} onblur={commit} />
	{:else}
		<button bind:this={trigger} type="button" class="atelier-editable-label__trigger" aria-label={`${ariaLabel}: ${value || placeholder}. Press Enter to edit`} title="Click or press Enter to edit" {disabled} onclick={beginEdit} onkeydown={handleTriggerKeydown}>
			<span class="atelier-editable-label__value" data-placeholder={value ? undefined : true}>{value || placeholder}</span>
		</button>
	{/if}
</span>
