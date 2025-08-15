<script>
  import Dots from "./theThreeDots.svelte"

  let {name, pfp, presence, duration} = $props()
</script>

<div class="flex-container">
  <div class="picture">
    <img src={pfp} class="picture">
  </div>
  <div class="friend-text">

    <h3>{name}</h3>
    <p>{presence}, {duration}</p>
  </div>

  <div class="dots">
    <Dots/>
  </div>

</div>

<style>
  h3 {
    color: var(--text)
  }
  p {color: var(--flamingo)}
  
  .flex-container {
    display: flex;
    align-items: center;
    flex: 1 1 15em;
    min-width: 10em;
    max-height: 4em;

    gap: 1em;

    border-width: 0.2em;
    border-style: solid;
    border-color: var(--overlay0);
    border-radius: 1em;
    
    padding: 1em;

    line-height: 0.2em;

    background-color: var(--surface0)
  }

  .dots {
    margin-left: auto;
  }
  
  .picture {
    height: 100%;
    width: auto;
  }

  img {
    border-radius: 50%;
  /*display: block;  Remove inline spacing */

  }
</style>
