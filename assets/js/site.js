var app = new Vue({
    el: '#app',

    data: {
        link: '',
        format: 'video',
        quality: 'high',
        count: 50,

        showModal: false,
        feedLink: '',

        // Server variables
        featureLevel: 0,
        userId: '',
    },

    methods: {
        submit: function() {
            var vm = this;

            if (vm.link === '') {
                return;
            }

            axios.post('/api/create', {
                url: this.link,
                format: this.format,
                quality: this.quality,
                page_size: this.count,
            }).then(function(response) {
                vm.feedLink = vm.formatLink(response.data.id);
                vm.showModal = true;
                vm.link = '';
            }).catch(vm.httpError);
        },

        httpError: function(error) {
            try {
                this.showError(error.response.data.error);
            } catch (e) {
                console.error(e);
                this.showError(error.message);
            }
        },

        showError: function(msg) {
            alert(msg);
        },

        formatLink: function(id) {
            if (location.port === '80' || location.port === '443') {
                return location.protocol + '//' + location.hostname + '/' + id;
            } else {
                return location.protocol + '//' + location.host + '/' + id;
            }
        },

        copyLink: function() {
            if (!this.showModal || !this.canCopy) {
                return
            }

            this.$refs.output.select();

            if (!document.execCommand('copy')) {
                self.showError('Can\'t copy... Something went wrong...');
            }
        }
    },

    computed: {
        locked: function() {
            return this.featureLevel === 0;
        },

        isMobile: function() {
            return /iPhone|iPad|iPod|Android/i.test(navigator.userAgent);
        },

        canCopy: function() {
            try {
                return document.queryCommandSupported('copy') && !this.isMobile;
            } catch (e) {
                return false;
            }
        }
    },

    mounted: function() {
        var vm = this;
        window.addEventListener('keydown', function(event) {
            // ESC handler
            if (event.keyCode === 27 && vm.showModal) {
                vm.showModal = false;
            }
        });
    }
});