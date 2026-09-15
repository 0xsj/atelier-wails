<script lang="ts">
	import type { Snippet } from 'svelte';
	import { Dialog as Primitive } from 'bits-ui';
	import './overlays.css';

	type DialogSize = 'sm' | 'md' | 'lg';

	interface Props {
		title: string;
		description?: string;
		trigger: Snippet;
		children: Snippet;
		open?: boolean;
		size?: DialogSize;
		class?: string;
	}

	let { title, description, trigger, children, open = $bindable(false), size = 'md', class: className }: Props = $props();
</script>

<Primitive.Root bind:open>
	<Primitive.Trigger class="atelier-dialog__trigger">
		{@render trigger()}
	</Primitive.Trigger>
	<Primitive.Portal>
		<Primitive.Overlay class="atelier-dialog__overlay" />
		<Primitive.Content class="atelier-dialog__content {className ?? ''}" data-size={size}>
			<div class="atelier-dialog__header">
				<div class="atelier-dialog__heading">
					<Primitive.Title class="atelier-dialog__title">{title}</Primitive.Title>
					{#if description}<Primitive.Description class="atelier-dialog__description">{description}</Primitive.Description>{/if}
				</div>
				<Primitive.Close class="atelier-dialog__close" aria-label="Close dialog">×</Primitive.Close>
			</div>
			<div class="atelier-dialog__body">{@render children()}</div>
		</Primitive.Content>
	</Primitive.Portal>
</Primitive.Root>
