<script>
  let {tabs, current = $bindable()} = $props()
</script>

<nav class="navigation">
  {#each tabs as tab}
    <button class="tab-button {current == tab ? "active" : ""}"
      onclick = {() => {current = tab}}>{tab}</button>
  {/each}
</nav>

<style>
/* Base tab button styles */
        .tab-button {
            display: inline-flex;
            align-items: center;
            justify-content: center;
            padding: 0.75rem 1.5rem;
            border: 2px solid transparent;
            background: transparent;
            color: var(--subtext1);
            font-size: 0.875rem;
            font-weight: 500;
            text-decoration: none;
            cursor: pointer;
            transition: all 0.2s ease;
            border-radius: 6px;
            position: relative;
            user-select: none;
            white-space: nowrap;
        }

        /* Hover states */
        .tab-button:hover {
            color: var(--text);
        }

        /* Tab containers */
        .navigation {
            display: flex;
            gap: 0;
            border-bottom: 2px solid #e2e8f0;
            margin-bottom: 1.5rem;
        }

        /* Style 1: Underline tabs */
        .tab-button {
            border-radius: 0;
            border-bottom: 3px solid transparent;
            padding-bottom: calc(0.75rem - 1px);
            margin-bottom: -2px;
        }

        .tab-button:hover {
            color: var(--text);
          border-bottom-color: var(--text);
        }

        .active {
            color: var(--mauve);
            border-bottom-color: var(--mauve);

            border-left-color: transparent;
            border-right-color: transparent;
            border-top-color: transparent;
          }
        </style>
