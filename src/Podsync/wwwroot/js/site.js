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
                    var text = '';

                    try {
                        var json = JSON.parse(xhr.responseText);
                        $.each(json, function (key, value) {
                            text += value + '\r\n';
                        });
                    } catch (e) {
                        text = xhr.responseText;
                    } 

                    err(text);
                } else {
                    // Generic error
                    err('Server sad \'' + error + '\': ' + xhr.responseText);
                }
            }
        });
    }

    function displayLink(link) {
        showModal(link);
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

    function getQuality() {
        var isAudio = $('#audio-format').hasClass('selected-option');
        var isWorst = $('#worst-quality').hasClass('selected-option');

        if (isAudio) {
            return isWorst ? 'AudioLow' : 'AudioHigh';
        } else {
            return isWorst ? 'VideoLow' : 'VideoHigh';
        }
    }

    /* Modal */

    function closeModal() {
        $('#modal').hide();
        $('#url-input').val('');
        $('.main').show();
    }

    function showModal(url) {
        // Hide main block on screen
        $('.main').hide();

        // Set input URL
        $('#output-input').val(url);

        // Update 'Open' button link
        $('#modal-open').attr('href', url);

        // Show dialog itself
        $('#modal').show();

        // Select modal output text
        $('#output-input').select();
    }

    function copyLink() {
        $('#output-input').select();
        if (!document.execCommand('copy')) {
            err('Can\'t copy... Something went wrong...');
        }
    }

    function canCopy() {
        try {
            return document.queryCommandSupported('copy');
        } catch (e) {
            return false;
        }
    }

    /*
        Attach handlers
    */

    $('#get-link').click(function(e) {
        var url = $('#url-input').val();
        createFeed({ url: url, quality: getQuality() }, displayLink);
        e.preventDefault();
    });

    $('#url-input').keyup(function (e) {
        // 'Enter' click
        if (e.keyCode === 13) {
            $('#get-link').click();
        }
    });

    $('#video-format, #audio-format').click(formatSwith);
    $('#best-quality, #worst-quality').click(qualitySwitch);

    $('#close-modal').click(closeModal);
    $('#modal-copy').click(copyLink);

    if (!canCopy()) {
        $('#modal-copy').hide();
    }

    $('body').on('keydown', function (e) {
        // ESC
        if ($('#modal').is(':visible') && e.keyCode === 27) {
            $('#close-modal').click();
        }
        e.stopPropagation();
    });
});