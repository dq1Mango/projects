<script>
  let {name, presence, duration} = $props()
</script>

<div class="flex-container">
  <h3>{name}</h3>
  <p>{presence}, {duration}</p>
</div>

<style>
  h3 {
    color: var(--text)
  }
  p {color: var(--flamingo)}

  .flex-container {
    outline-width: 5px;
    outline-style: solid;
    outline-color: var(--maroon);

    margin: 10px ;
    padding-left: 10px;
}
</style>
