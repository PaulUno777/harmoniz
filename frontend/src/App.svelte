<script lang="ts">
  import logo from './assets/images/logo-universal.png'
  import {Greet} from '../wailsjs/go/main/App.js'

  let resultText: string = "Please enter your name below 👇"
  let name: string

  function greet(): void {
    Greet(name).then(result => resultText = result)
  }
</script>

<main class="flex flex-col items-center justify-center min-h-screen bg-slate-50 text-slate-900 font-sans">
  
  <img 
    alt="Wails logo" 
    src="{logo}" 
    class="block w-1/2 max-w-xs mb-8 transition-transform hover:scale-105 duration-500"
  >

  <div class="h-6 mb-6 text-lg font-medium text-slate-600">
    {resultText}
  </div>

  <div class="flex gap-4">
    <input 
      autocomplete="off" 
      bind:value={name} 
      type="text"
      placeholder="Type your name..."
      class="h-10 px-4 rounded-md border-none outline-none bg-slate-200 focus:bg-white focus:ring-2 focus:ring-sky-500 transition-all antialiased"
    />
    
    <button 
      class="h-10 px-6 rounded-md bg-slate-800 text-white font-semibold cursor-pointer 
             hover:bg-linear-to-t hover:from-slate-200 hover:to-slate-50 hover:text-slate-900 
             active:scale-95 transition-all shadow-md"
      on:click={greet}
    >
      Greet
    </button>
  </div>

</main>

