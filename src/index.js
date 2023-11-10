import './style.css';
import '@riotjs/hot-reload';
import { register, component } from 'riot';

// routes
import { Route, Router } from '@riotjs/route';
register('router', Router);
register('route', Route);

// views
import Overview from './views/overview.riot';
import List from './views/list.riot';
import Detail from './views/detail.riot';
register('overview', Overview);
register('list', List);
register('detail', Detail);

// app
import App from './app.riot';
const mountApp = component(App);
mountApp(document.getElementById('root'));

