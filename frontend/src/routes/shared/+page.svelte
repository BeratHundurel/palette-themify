<script lang="ts">
	import { goto } from '$app/navigation';
	import { navigating, page } from '$app/state';
	import { resolve } from '$app/paths';
	import { SvelteURLSearchParams } from 'svelte/reactivity';
	import toast, { Toaster } from 'svelte-french-toast';

	import { generateOverridable, type EditorThemeType } from '$lib/api/theme';
	import { detectThemeAppearance, detectThemeType } from '$lib/colorUtils';
	import { hydrateThemeExportResponse } from '$lib/components/toolbar/theme-export/session';
	import { popoverStore } from '$lib/stores/popovers.svelte';
	import { appStore } from '$lib/stores/app/store.svelte';
	import BrandLinks from '$lib/components/ui/BrandLinks.svelte';
	import type { SharedItem, SharedItemSort } from '$lib/types/shared';
	import type { Theme } from '$lib/types/theme';

	type SharedPageData = {
		items: SharedItem[];
		errorMessage: string;
	};

	let { data }: { data: SharedPageData } = $props();

	let searchQuery = $state('');
	let sortBy = $state<SharedItemSort>('newest');

	const items = $derived(data.items);
	const errorMessage = $derived(data.errorMessage);
	const isLoading = $derived.by(() => {
		const nav = navigating;
		return !!nav?.to && nav.to.url.pathname === page.url.pathname;
	});

	function buildSharedQuery(query: string, sort: SharedItemSort): string {
		const params = new SvelteURLSearchParams();
		const trimmedQuery = query.trim();
		if (trimmedQuery) params.set('q', trimmedQuery);
		if (sort !== 'newest') params.set('sort', sort);
		const queryString = params.toString();
		return queryString ? `?${queryString}` : '';
	}

	async function refreshSharedItems(query: string, sort: SharedItemSort, force = false): Promise<void> {
		const queryString = buildSharedQuery(query, sort);
		let target = resolve('/shared');
		target += queryString;
		const current = `${page.url.pathname}${page.url.search}`;

		if (target === current && !force) return;

		await goto(target, {
			replaceState: true,
			keepFocus: true,
			noScroll: true,
			invalidateAll: true
		});
	}

	$effect(() => {
		searchQuery = (page.url.searchParams.get('q') ?? '').trim();
		const sortParam = page.url.searchParams.get('sort');
		sortBy = sortParam === 'oldest' || sortParam === 'name' ? sortParam : 'newest';
	});

	$effect(() => {
		const query = searchQuery.trim();
		const sort = sortBy;
		const currentQuery = (page.url.searchParams.get('q') ?? '').trim();
		const currentSortParam = page.url.searchParams.get('sort');
		const currentSort = currentSortParam === 'oldest' || currentSortParam === 'name' ? currentSortParam : 'newest';

		if (query === currentQuery && sort === currentSort) return;

		const timeout = setTimeout(() => {
			void refreshSharedItems(query, sort);
		}, 220);

		return () => clearTimeout(timeout);
	});

	async function importAsTheme(item: SharedItem) {
		if (item.kind === 'palette') {
			if (!item.palette || item.palette.length === 0) {
				toast.error('This shared palette has no colors.');
				return;
			}

			appStore.resetThemeExportSession();
			appStore.state.colors = [];
			appStore.state.image = null;
			appStore.state.imageLoaded = false;
			appStore.state.themeExport.backupColors = item.palette.map((color) => ({ hex: color.hex }));
			await goto(resolve('/'));
			popoverStore.state.current = 'themeExport';
			return;
		}

		if (!item.theme || typeof item.theme !== 'object') {
			toast.error('This shared theme is not importable.');
			return;
		}

		const theme = item.theme as Theme;
		const editorType = (item.editorType as EditorThemeType | undefined) ?? detectThemeType(theme);
		const appearance = detectThemeAppearance(theme);

		try {
			const response = await generateOverridable(theme, null, editorType, appearance);

			appStore.state.colors = [];
			appStore.state.image = null;
			appStore.state.imageLoaded = false;
			hydrateThemeExportResponse(response, editorType, appearance);

			await goto(resolve('/'));
			popoverStore.state.current = 'themeExport';
		} catch {
			toast.error('Could not import this shared theme.');
		}
	}

	function formatSharedDate(value: string): string {
		const date = new Date(value);
		if (Number.isNaN(date.getTime())) return 'Unknown date';
		return date.toLocaleString('en-US', {
			month: 'short',
			day: 'numeric',
			year: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	}
</script>

<Toaster />

<div class="relative min-h-svh bg-zinc-950 text-zinc-100">
	<enhanced:img
		src="../../lib/assets/palette.jpg"
		alt="Palette"
		class="absolute top-0 left-0 h-full w-full object-cover"
	/>

	<div class="absolute top-0 left-0 z-10 h-full w-full bg-black/60"></div>

	<div class="relative z-40 mx-auto w-full max-w-6xl px-4 py-8 md:px-6">
		<div class="mb-6 flex flex-wrap items-center justify-between gap-3">
			<div>
				<p class="text-brand text-xs font-semibold tracking-[0.14em] uppercase">Community</p>
				<h1 class="mt-1 text-2xl font-bold md:text-3xl">Shared Themes and Palettes</h1>
				<p class="mt-1 text-sm text-zinc-400">Browse public creations and import any item as a theme.</p>
			</div>
			<BrandLinks href="/">Back</BrandLinks>
		</div>

		<div class="mb-5 grid gap-3 md:grid-cols-[1fr_180px_auto]">
			<input
				type="search"
				bind:value={searchQuery}
				placeholder="Search by name"
				class="focus:border-brand/70 w-full rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm outline-none"
			/>
			<select
				bind:value={sortBy}
				class="focus:border-brand/70 rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm outline-none"
			>
				<option value="newest">Newest shared</option>
				<option value="oldest">Oldest shared</option>
				<option value="name">Name A-Z</option>
			</select>
			<button
				type="button"
				onclick={() => void refreshSharedItems(searchQuery, sortBy, true)}
				class="bg-brand rounded-md px-4 py-2 text-sm font-semibold text-zinc-900"
			>
				Refresh
			</button>
		</div>

		{#if isLoading}
			<p class="text-sm text-zinc-400">Loading shared items...</p>
		{:else if errorMessage}
			<p class="rounded-md border border-red-500/40 bg-red-500/10 px-3 py-2 text-sm text-red-200">{errorMessage}</p>
		{:else if items.length === 0}
			<p class="rounded-md border border-zinc-800 bg-zinc-900/60 px-4 py-6 text-center text-sm text-zinc-400">
				No shared items found for this filter.
			</p>
		{:else}
			<ul class="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
				{#each items as item (item.id)}
					<li class="rounded-xl border border-zinc-700 bg-zinc-900/80 p-3">
						<div class="mb-2 flex items-center justify-between gap-2">
							<p class="truncate text-sm font-semibold" title={item.name}>{item.name}</p>
							<span class="rounded-full border border-zinc-600 px-2 py-0.5 text-[10px] font-semibold uppercase">
								{item.kind}
							</span>
						</div>

						<div class="mb-2 flex flex-wrap gap-1">
							{#if item.palette.length > 0}
								{#each item.palette.slice(0, 12) as color, index (item.id + '-' + index)}
									<div
										title={color.hex}
										class="h-6 w-6 rounded-md border border-zinc-700"
										style={`background-color: ${color.hex}`}
									></div>
								{/each}
							{:else}
								<span class="text-xs text-zinc-500">No palette preview</span>
							{/if}
						</div>

						<p class="mb-3 text-xs text-zinc-500">Shared {formatSharedDate(item.sharedAt)}</p>

						<button
							type="button"
							onclick={() => void importAsTheme(item)}
							class="border-brand/40 bg-brand/15 text-brand hover:bg-brand/25 w-full rounded-md border px-3 py-2 text-xs font-semibold"
						>
							Import as Theme
						</button>
					</li>
				{/each}
			</ul>
		{/if}
	</div>
</div>
