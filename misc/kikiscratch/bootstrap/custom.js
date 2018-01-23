
$( document ).ready(function() {
  console.log( "ready!" );
  $('#submit-post').click(function(e){
    e.preventDefault();
    console.log("posting")
    $.ajax({
         url: "http://127.0.0.1:8003/letterhtml",
         type: "POST",//type of posting the data
         data: JSON.stringify({"data":editor.getContent(),"kind":"post","reply-to":"TODO","privacy":"FRIENDS/PUBLIC/PERSONAL"}),
         success: function (data) {
           $('#exampleModal').modal('hide');
           editor.setContent("");
           $.snackbar({content: "This is my awesome snackbar!"});
         },
         error: function(xhr, ajaxOptions, thrownError){
          $.snackbar({content: "Could not post: " + xhr.statusText});
          console.log(thrownError);
        },
         timeout : 500//timeout of the ajax call
    });
  
  });
  
  
});
