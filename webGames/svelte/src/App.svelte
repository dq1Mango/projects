<script>
  import Gopher from "/src/assets/gopher.jpg"
  import Friend from './lib/Friend.svelte'  
  import Tabs from "./lib/Tabs.svelte"
  
  let friends = $state([
    {name : "sarah", pfp: Gopher, presence : "online", duration : "now"},
    {name : "alex", pfp: Gopher, presence : "offline", duration : "2 hours"}
  ])

  const tabs = ["friends", "availible", "active"]

  let current = $state("friends")
  let showFriends = $derived(current == "friends")
  let showAvailible = $derived(current == "availible")
  let showActive = $derived(current == "active")

  $inspect(current)

</script>

<main>
  <div class="headers">
    <h1>
      Dashboard
    </h1>
    <p>
      Some stupid text telling you what this page does
    </p>
  </div>

  <div>
    <Tabs tabs = {tabs} bind:current = {current}/>
  </div>

  {#if showFriends}
    <div class="tab">
      <h2>Your Friends ({friends.length}):</h2>
      <div class="friend-box"> 
      {#each friends as friend}
        <Friend {...friend} />
      {/each}
      </div>
    </div>
  {/if}

{#if showAvailible}
  <h2>Availible Games</h2>
{/if}




</main>

<style>
  /* your styles go here */
  main {
    margin: 1em;
  }

  .headers {
  }

  .tab {
  }

  .friend-box {
    display: flex;
    flex-direction: row;
    flex-wrap: wrap;
    gap: 1em;

  }

  h2 {
    color: var(--mauve);
  }

</style>
