// Write your Javascript code.

$(function () {

    function err(msg) {
        alert(msg);
    }
    
    function createFeed(data, done) {
        if (!data.url) {
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

    /*
        Tooltips
    */

    $(document).on('mouseenter', 'i', function() {
        var title = $(this).attr('title');
        if (!title) {
            return;
        }
        $(this).data('tipText', title).removeAttr('title');
        $('<p class="tooltip"></p>').text(title).appendTo('body').fadeIn('fast');
    });

    $(document).on('mouseleave', 'i', function() {
        var text = $(this).data('tipText');
        $(this).attr('title', text);
        $('.tooltip').remove();
    });

    $(document).on('mousemove', 'i', function(e) {
        var x = e.pageX + 10;
        var y = e.pageY + 5;
        $('.tooltip').css({ top: y, left: x });
    });

    /*
        Control panel
    */

    function lockControls(lock) {
        if (lock) {
            $('#control-panel').addClass('locked');
            $('#control-icon')
                .removeClass('fa-wrench')
                .addClass('fa-question-circle master-tooltip')
                .attr('title', 'This features are available for patrons only (please, login with your Patreon account). You may support us and unlock this features');
        } else {
            $('#control-panel')
                .removeClass('locked');
            $('#control-icon')
                .removeClass('fa-question-circle master-tooltip')
                .addClass('fa-wrench')
                .removeAttr('title');
        }
    }

    function isLocked() {
        return $('#control-panel').hasClass('locked');
    }

    /*
        Handlers
    */

    function formatSwith() {
        if (isLocked()) {
            return;
        }

        $('#video-format, #audio-format').toggleClass('selected-option');
    }

    function qualitySwitch() {
        if (isLocked()) {
            return;
        }

        $('#best-quality, #worst-quality').toggleClass('selected-option');
    }

    /*
        Attach handlers
    */

    $('#get-link').click(function(e) {
        var url = $('#url-input').val();
        createFeed({ url: url, quality: 'VideoHigh' }, displayLink);
        e.preventDefault();
    });

    $('#video-format, #audio-format').click(formatSwith);
    $('#best-quality, #worst-quality').click(qualitySwitch);

    lockControls(true);
});