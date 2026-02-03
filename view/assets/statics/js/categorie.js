(function(){
      const labels = document.querySelectorAll('.categories-row label.cat-btn');
      if(!labels) return;
      labels.forEach(lbl => {
        const inp = lbl.querySelector('input[type="checkbox"]');
        if(inp && inp.checked) lbl.classList.add('active');
        lbl.addEventListener('click', function(e){
          setTimeout(()=>{
            if(inp && inp.checked) lbl.classList.add('active'); else lbl.classList.remove('active');
          }, 10);
        });
      });
    })();