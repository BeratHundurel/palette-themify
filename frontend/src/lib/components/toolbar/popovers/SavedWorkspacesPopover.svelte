<script lang="ts">
	import { cn } from '$lib/utils';
	import { appStore } from '$lib/stores/app.svelte';
	import { popoverStore } from '$lib/stores/popovers.svelte';
	import { shareWorkspace, removeWorkspaceShare } from '$lib/api/workspace';
	import toast from 'svelte-french-toast';

	let shareModal = $state<{ workspaceId: string; workspaceName: string; shareUrl: string } | null>(null);
	let sharingWorkspaceId = $state<string | null>(null);

	async function handleWorkspaceLoad(workspaceId: string) {
		const workspace = appStore.state.savedWorkspaces.find((w) => w.id === workspaceId);
		if (workspace) {
			await appStore.loadWorkspace(workspace);
			popoverStore.close('workspaces');
		}
	}

	async function handleWorkspaceDelete(workspaceId: string, workspaceName: string) {
		if (confirm(`Are you sure you want to delete "${workspaceName}"?`)) {
			await appStore.deleteWorkspace(workspaceId);
		}
	}

	async function handleShare(workspaceId: string, workspaceName: string) {
		sharingWorkspaceId = workspaceId;
		const toastId = toast.loading('Generating share link...');

		try {
			const result = await shareWorkspace(workspaceId);

			const workspace = appStore.state.savedWorkspaces.find((w) => w.id === workspaceId);
			if (workspace) {
				workspace.shareToken = result.shareToken;
			}

			shareModal = {
				workspaceId,
				workspaceName,
				shareUrl: result.shareUrl
			};

			toast.success('Share link created!', { id: toastId });
		} catch {
			toast.error('Could not create a share link. Please try again.', {
				id: toastId
			});
		} finally {
			sharingWorkspaceId = null;
		}
	}

	async function handleCopyLink() {
		if (!shareModal) return;

		try {
			await navigator.clipboard.writeText(shareModal.shareUrl);
			toast.success('Link copied to clipboard!');
		} catch {
			toast.error('Could not copy the link. Please try again.');
		}
	}

	async function handleRemoveShare() {
		if (!shareModal) return;

		const workspaceId = shareModal.workspaceId;
		const toastId = toast.loading('Removing share link...');

		try {
			await removeWorkspaceShare(workspaceId);

			const workspace = appStore.state.savedWorkspaces.find((w) => w.id === workspaceId);
			if (workspace) {
				workspace.shareToken = null;
			}

			toast.success('Share link removed', { id: toastId });
			shareModal = null;
		} catch {
			toast.error('Could not remove the share link. Please try again.', {
				id: toastId
			});
		}
	}

	function openShareModal(workspaceId: string, workspaceName: string, shareToken: string) {
		shareModal = {
			workspaceId,
			workspaceName,
			shareUrl: `http://localhost:5173/?share=${shareToken}`
		};
	}
</script>

<div
	class={cn(
		'palette-dropdown-base w-80',
		popoverStore.state.direction === 'right' ? 'left-full ml-2' : 'right-full mr-2'
	)}
	style={`min-width: 260px; ${popoverStore.state.direction === 'right' ? 'left: calc(100% + 0.5rem);' : 'right: calc(100% + 0.5rem);'}`}
	role="dialog"
	aria-labelledby="saved-workspaces-title"
	tabindex="-1"
>
	<h3 id="saved-workspaces-title" class="text-brand mb-3 text-xs font-medium">Saved Workspaces</h3>
	<div class="scrollable-content custom-scrollbar max-h-72 overflow-y-auto">
		{#if appStore.state.savedWorkspaces.length === 0}
			<div class="flex flex-col items-center justify-center py-12 text-center">
				<svg class="mb-3 h-12 w-12 text-zinc-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="1.5"
						d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"
					/>
				</svg>
				<p class="text-sm text-zinc-400">No saved workspaces yet</p>
				<p class="mt-1 text-xs text-zinc-500">Save your current workspace to see it here</p>
			</div>
		{:else}
			<ul class="flex flex-col gap-3">
				{#each appStore.state.savedWorkspaces as item (item.id)}
					<li
						class="hover:border-brand/50 group relative overflow-hidden rounded-lg border border-zinc-600 bg-zinc-800/50 transition-[background-color,border-color,box-shadow] duration-300 hover:bg-white/5"
					>
						<div class="flex gap-3 p-3">
							<!-- Thumbnail -->
							<div class="relative shrink-0">
								<img
									src={item.imageData}
									alt={item.name}
									class="h-20 w-20 rounded-md border border-zinc-700/50 object-cover shadow-md transition-transform duration-300 group-hover:scale-105"
								/>
								{#if item.shareToken}
									<div class="absolute -top-1 -right-1 rounded-full bg-blue-500 p-1 shadow-lg" title="Shared">
										<svg class="h-3 w-3 text-white" fill="currentColor" viewBox="0 0 20 20">
											<path
												d="M15 8a3 3 0 10-2.977-2.63l-4.94 2.47a3 3 0 100 4.319l4.94 2.47a3 3 0 10.895-1.789l-4.94-2.47a3.027 3.027 0 000-.74l4.94-2.47C13.456 7.68 14.19 8 15 8z"
											/>
										</svg>
									</div>
								{/if}
							</div>

							<!-- Content -->
							<div class="flex min-w-0 flex-1 flex-col justify-between">
								<!-- Header -->
								<div class="mb-2">
									<h4 class="text-brand mb-1 truncate font-mono text-sm font-semibold" title={item.name}>
										{item.name}
									</h4>
									<div class="flex items-center gap-3 text-xs text-zinc-400">
										<span class="flex items-center gap-1">
											<svg class="h-3.5 w-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path
													stroke-linecap="round"
													stroke-linejoin="round"
													stroke-width="2"
													d="M7 21a4 4 0 01-4-4V5a2 2 0 012-2h4a2 2 0 012 2v12a4 4 0 01-4 4zm0 0h12a2 2 0 002-2v-4a2 2 0 00-2-2h-2.343M11 7.343l1.657-1.657a2 2 0 012.828 0l2.829 2.829a2 2 0 010 2.828l-8.486 8.485M7 17h.01"
												/>
											</svg>
											{item.colors?.length || 0}
										</span>
										<span class="flex items-center gap-1">
											<svg class="h-3.5 w-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path
													stroke-linecap="round"
													stroke-linejoin="round"
													stroke-width="2"
													d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"
												/>
											</svg>
											{item.selectors?.length || 0}
										</span>
									</div>
								</div>

								<!-- Actions -->
								<div class="flex items-center gap-2">
									<button
										class="text-brand hover:bg-brand/10 flex items-center gap-1.5 rounded-md py-1.5 ps-0 pe-2.5 text-xs font-medium transition-[transform,background-color] hover:scale-105"
										onclick={() => handleWorkspaceLoad(item.id)}
										type="button"
										title="Load workspace"
									>
										<svg class="h-3.5 w-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path
												stroke-linecap="round"
												stroke-linejoin="round"
												stroke-width="2"
												d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"
											/>
										</svg>
										Load
									</button>

									<button
										class={cn(
											'flex items-center gap-1 rounded-md p-1.5 transition-[transform,background-color,color] hover:scale-110 hover:bg-blue-500/10',
											item.shareToken ? 'text-blue-400 hover:text-blue-300' : 'text-zinc-500 hover:text-zinc-400'
										)}
										onclick={() =>
											item.shareToken
												? openShareModal(item.id, item.name, item.shareToken!)
												: handleShare(item.id, item.name)}
										type="button"
										title={item.shareToken ? 'View share link' : 'Share workspace'}
										disabled={sharingWorkspaceId === item.id}
									>
										<svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path
												stroke-linecap="round"
												stroke-linejoin="round"
												stroke-width="2"
												d="M8.684 13.342C8.886 12.938 9 12.482 9 12c0-.482-.114-.938-.316-1.342m0 2.684a3 3 0 110-2.684m0 2.684l6.632 3.316m-6.632-6l6.632-3.316m0 0a3 3 0 105.367-2.684 3 3 0 00-5.367 2.684zm0 9.316a3 3 0 105.368 2.684 3 3 0 00-5.368-2.684z"
											/>
										</svg>
									</button>

									<button
										class="flex items-center gap-1 rounded-md p-1.5 text-zinc-500 transition-[transform,background-color,color] hover:scale-110 hover:bg-red-500/10 hover:text-red-400"
										onclick={() => handleWorkspaceDelete(item.id, item.name)}
										type="button"
										title="Delete workspace"
									>
										<svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path
												stroke-linecap="round"
												stroke-linejoin="round"
												stroke-width="2"
												d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
											/>
										</svg>
									</button>
								</div>
							</div>
						</div>

						<!-- Footer - Creation Date -->
						<div class="group-hover:border-brand/50 group border-t border-zinc-600 bg-zinc-900/50 px-3 py-1.5">
							<span class="flex items-center gap-1.5 text-xs text-zinc-500">
								<svg class="h-3 w-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path
										stroke-linecap="round"
										stroke-linejoin="round"
										stroke-width="2"
										d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"
									/>
								</svg>
								{new Date(item.createdAt).toLocaleDateString('en-US', {
									month: 'short',
									day: 'numeric',
									year: 'numeric'
								})}
							</span>
						</div>
					</li>
				{/each}
			</ul>
		{/if}
	</div>
</div>

{#if shareModal}
	<div class="fixed inset-0 z-[100] flex items-center justify-center bg-black/50">
		<div class="share-modal-content border-brand/50 w-full max-w-md rounded-lg border bg-zinc-900 p-6">
			<h2 class="text-brand mb-4 text-lg font-semibold">Share Workspace</h2>

			<p class="mb-4 text-sm text-zinc-400">
				Anyone with this link can view your workspace. They won't be able to modify it.
			</p>

			<div class="mb-4 flex gap-2">
				<input
					type="text"
					readonly
					value={shareModal.shareUrl}
					class="flex-1 rounded border border-zinc-700 bg-zinc-800 px-3 py-2 text-sm text-zinc-200"
				/>
				<button
					onclick={handleCopyLink}
					class="bg-brand hover:bg-brand-hover cursor-pointer rounded px-4 py-2 text-sm font-medium text-zinc-900 transition"
				>
					Copy
				</button>
			</div>

			<div class="flex justify-between gap-2">
				<button
					onclick={handleRemoveShare}
					class="cursor-pointer rounded bg-red-500 px-4 py-2 text-sm font-medium text-white transition hover:bg-red-600"
				>
					Remove Share
				</button>
				<button
					onclick={() => (shareModal = null)}
					class="cursor-pointer rounded bg-zinc-700 px-4 py-2 text-sm font-medium text-zinc-200 transition hover:bg-zinc-600"
				>
					Close
				</button>
			</div>
		</div>
	</div>
{/if}
