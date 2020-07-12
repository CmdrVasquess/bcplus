const protoApp = new Vue({
    el: "main",
    data: {},
    methods: {
	onMsg(evt) {
	    console.log("Screen proto received: "+evt);
	}
    },
    mounted: function() {
	const app = this;
    }
});
