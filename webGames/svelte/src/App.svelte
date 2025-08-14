<script>
  import Friend from './lib/Friend.svelte'  
  
  let friends = $state([
    {name : "sarah", presence : "online", duration : "now"},
    {name : "alex", presence : "offline", duration : "2 hours"}
  ])
</script>

<main>
  <div>
    <h1>
      Dashboard
    </h1>
    <p>
      Some stupid text telling you what this page does
    </p>
  </div>
  <div>
    {#each friends as friend}
      <Friend {...friend} />
    {/each}
  </div>
</main>
