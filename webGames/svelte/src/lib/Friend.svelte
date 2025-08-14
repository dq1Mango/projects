<script>
  let {name, presence, duration} = $props()
</script>

<div class="flex-container">
  <div class="friend-text">

    <h3>{name}</h3>
    <p>{presence}, {duration}</p>
  </div>
  <div>
  </div>

</div>

<style>
  h3 {
    color: var(--text)
  }
  p {color: var(--flamingo)}
  
  .flex-container {
    display: flex;

    border-width: 0.2em;
    border-style: solid;
    border-color: var(--sapphire);
    border-radius: 1em;
    
    padding: 1em;

    line-height: 0.2em;
}
</style>
