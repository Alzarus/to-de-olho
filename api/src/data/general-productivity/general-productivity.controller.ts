import {
  Controller,
  Get,
  Post,
  Body,
  Patch,
  Param,
  Delete,
} from '@nestjs/common';
import { GeneralProductivityService } from './general-productivity.service';
import { CreateGeneralProductivityDto } from './dto/create-general-productivity.dto';
import { UpdateGeneralProductivityDto } from './dto/update-general-productivity.dto';

@Controller('general-productivity')
export class GeneralProductivityController {
  constructor(
    private readonly generalProductivityService: GeneralProductivityService,
  ) {}

  @Post()
  async create(
    @Body() data: CreateGeneralProductivityDto | CreateGeneralProductivityDto[],
  ) {
    if (Array.isArray(data)) {
      return this.generalProductivityService.createMany(data);
    } else {
      return this.generalProductivityService.create(data);
    }
  }

  @Get()
  findAll() {
    return this.generalProductivityService.findAll();
  }

  @Get('latest')
  findLatest() {
    return this.generalProductivityService.findLatest();
  }

  @Get(':id')
  findOne(@Param('id') id: string) {
    return this.generalProductivityService.findOne(+id);
  }

  @Patch(':id')
  update(
    @Param('id') id: string,
    @Body() updateGeneralProductivityDto: UpdateGeneralProductivityDto,
  ) {
    return this.generalProductivityService.update(
      +id,
      updateGeneralProductivityDto,
    );
  }

  @Delete(':id')
  remove(@Param('id') id: string) {
    return this.generalProductivityService.remove(+id);
  }
}
