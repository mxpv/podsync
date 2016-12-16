// Write your Javascript code.

$(function () {

    function err(msg) {
        alert(msg);
    }
    
    function createFeed(data, done) {
        if (!data.url) {
            err('Please fill the URL field first');
            return;
        }

        $.ajax({
            dataType: 'text',
            url: '/feed/create',
            method: 'POST',
            data: JSON.stringify(data),
            contentType: 'application/json; charset=utf-8',
            success: function (feedId) {
                var proto = $(location).attr('protocol');
                var host = $(location).attr('host');
                done(proto + '//' + host + '/feed/' + feedId);
            },
            error: function (xhr, status, error) {
                if (xhr.status === 400) {
                    // Bad request
                    var json = JSON.parse(xhr.responseText);

                    var text = '';
                    $.each(json, function (key, value) {
                        text += value + '\r\n';
                    });

                    err(text);
                } else {
                    // Generic error
                    err('Server sad \'' + error + '\': ' + xhr.responseText);
                }
            }
        });
    }

    function displayLink(link) {
        alert(link);
    }

    $('#get-link').click(function(e) {
        var url = $('#url-input').val();
        createFeed({ url: url, quality: 'VideoHigh' }, displayLink);
        e.preventDefault();
    });
});