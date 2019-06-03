import { BrowserModule } from '@angular/platform-browser';
import { NgModule } from '@angular/core';

import { AppRoutingModule } from './app-routing.module';
import { AppComponent } from './app.component';
import { IndexComponent } from './index/index.component';
import { InputComponent } from './input/input.component';
import {FormsModule} from '@angular/forms';
import {HttpClientModule} from '@angular/common/http';
import { OutputComponent } from './output/output.component';
import { FaqComponent } from './faq/faq.component';
import { FooterComponent } from './widgets/footer/footer.component';

@NgModule({
  declarations: [
    AppComponent,
    IndexComponent,
    InputComponent,
    OutputComponent,
    FaqComponent,
    FooterComponent
  ],
  imports: [
    BrowserModule,
    HttpClientModule,
    FormsModule,
    AppRoutingModule
  ],
  providers: [],
  bootstrap: [AppComponent]
})
export class AppModule { }
