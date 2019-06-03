import { Component } from '@angular/core';
import {Meta, Title} from '@angular/platform-browser';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.scss']
})
export class AppComponent {
  constructor(private titleService: Title, metaService: Meta) {
    titleService.setTitle('Podsync - Turn YouTube channels into podcast feeds');

    metaService.addTag({
      httpEquiv: 'content-type',
      content: 'text/html;charset=UTF-8',
    });

    metaService.addTag({
      name: 'description',
      content: 'Simple and free service that lets you listen to any YouTube or Vimeo channels, playlists or user videos in podcast format',
    });

    metaService.addTag({
      name: 'og:title',
      content: 'Podsync - turn YouTube channels into podcast feeds',
    });

    metaService.addTag({
      name: 'og:description',
      content: 'Simple and free service that lets you listen to any YouTube or Vimeo channels, playlists or user videos in podcast format',
    });

    metaService.addTag({
      name: 'og:locale',
      content: 'en_US',
    });

    metaService.addTag({
      name: 'og:image',
      content: '/assets/img/og_image.png',
    });
  }
}
