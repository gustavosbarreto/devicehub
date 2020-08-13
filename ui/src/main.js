import Vue from 'vue';

import Snackbar from '@/components/snackbar/Snackbar';

import App from './App';
import router from './router';
import store from './store';
import env from './env';
import './vee-validate';
import vuetify from './plugins/vuetify';

Vue.config.productionTip = false;

Vue.component('Snackbar', Snackbar);

Vue.use(require('vue-moment'));

Vue.use(env);

new Vue({
  vuetify,
  router,
  store,
  render: (h) => h(App),
}).$mount('#app');
