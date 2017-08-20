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
            url: '/api/create',
            method: 'POST',
            data: JSON.stringify(data),
            contentType: 'application/json; charset=utf-8',
            success: function (feedLink) {
                // HACK: remove quotes
                feedLink = JSON.parse(feedLink);
                done(feedLink);
            },
            error: function (xhr, status, error) {
                var text = '';

                try {
                    var json = JSON.parse(xhr.responseText);
                    if (json['error']) {
                        text = json['error'];
                    }
                } catch (e) {
                    text = xhr.responseText;
                }

                err(text);
            }
        });
    }

    function displayLink(link) {
        showModal(link);
    }

    /*
        Tooltips
    */

    if (!isMobile()) {
        $(document).on('mouseenter', 'i', function () {
            var title = $(this).attr('title');
            if (!title) {
                return;
            }

            $(this).data('tipText', title).removeAttr('title');
            $('<p class="tooltip"></p>').text(title).appendTo('body').fadeIn('fast');
        });

        $(document).on('mouseleave', 'i', function () {
            var text = $(this).data('tipText');
            $(this).attr('title', text);
            $('.tooltip').remove();
        });

        $(document).on('mousemove', 'i', function (e) {
            var x = e.pageX + 10;
            var y = e.pageY + 5;
            $('.tooltip').css({ top: y, left: x });
        });
    }

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

    function getFormat() {
        var isAudio = $('#audio-format').hasClass('selected-option');
        return isAudio ? 'audio' : 'video'
    }

    function getQuality() {
        var isWorst = $('#worst-quality').hasClass('selected-option');
        return isWorst ? 'low' : 'high';
    }

    function pageSwitch(evt) {
        if (isLocked()) {
            return;
        }

        $('#page-controls > a').removeClass('selected-option');
        $(evt.target).addClass('selected-option');

        getPageCount();
    }

    function getPageCount() {
        try {
            var text = $('#page-controls > a.selected-option').text();
            return parseInt(text);
        } catch (e) {
            return 50;
        }
    }

    $('#page-controls > a').click(pageSwitch);

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

    function isMobile() {
        return /iPhone|iPad|iPod|Android/i.test(navigator.userAgent);
    }

    function canCopy() {
        try {
            return document.queryCommandSupported('copy') && !isMobile();
        } catch (e) {
            return false;
        }
    }

    $('body').on('keydown', function (e) {
        // ESC
        if ($('#modal').is(':visible') && e.keyCode === 27) {
            $('#close-modal').click();
        }
        e.stopPropagation();
    });

    /*
        Attach handlers
    */

    $('#get-link').click(function(e) {
        var url = $('#url-input').val();
        createFeed({ url: url, format: getFormat(), quality: getQuality(), page_size: getPageCount() }, displayLink);
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
});