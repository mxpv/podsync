import { Component, OnInit } from '@angular/core';

@Component({
  selector: 'app-input',
  templateUrl: './input.component.html',
  styleUrls: ['./input.component.scss']
})
export class InputComponent implements OnInit {
  popupOpened: boolean;
  locked = false;
  format = 'video';
  quality = 'high';
  pageSize = 50;
  link = '';

  constructor() { }

  ngOnInit() {
  }

  submit() {
    console.log('input submit');
  }
}
